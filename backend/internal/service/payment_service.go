package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/florus/backend/internal/config"
	"github.com/florus/backend/internal/dto"
	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrPaymentNotFound     = errors.New("payment not found")
	ErrPaymentAlreadyPaid  = errors.New("payment already completed")
	ErrInvalidPaymentMethod = errors.New("invalid payment method")
	ErrPaymentFailed       = errors.New("payment creation failed")
)

type PaymentService interface {
	CreatePayment(orderCode string, method models.PaymentMethod) (*dto.CreatePaymentResponse, error)
	GetPaymentByOrderCode(orderCode string) (*models.Payment, error)
	HandleZaloPayCallback(data map[string]interface{}) error
	HandleMoMoCallback(data map[string]interface{}) error
	HandleVNPayCallback(params url.Values) error
}

type paymentService struct {
	db           *gorm.DB
	orderRepo    repository.OrderRepository
	paymentCfg   config.PaymentConfig
}

func NewPaymentService(db *gorm.DB, orderRepo repository.OrderRepository, paymentCfg config.PaymentConfig) PaymentService {
	return &paymentService{
		db:         db,
		orderRepo:  orderRepo,
		paymentCfg: paymentCfg,
	}
}

func (s *paymentService) CreatePayment(orderCode string, method models.PaymentMethod) (*dto.CreatePaymentResponse, error) {
	// Get order
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return nil, ErrOrderNotFound
	}

	// Check if payment already exists and is successful
	var existingPayment models.Payment
	if err := s.db.Where("order_id = ? AND status = ?", order.ID, models.PaymentStatusSuccess).First(&existingPayment).Error; err == nil {
		return nil, ErrPaymentAlreadyPaid
	}

	// Generate transaction ID
	transactionID := fmt.Sprintf("TXN%d%s", time.Now().Unix(), uuid.New().String()[:8])

	// Create payment record
	payment := &models.Payment{
		OrderID:       order.ID,
		TransactionID: transactionID,
		Method:        method,
		Amount:        order.Total,
		Status:        models.PaymentStatusPending,
	}

	// Generate payment URL based on method
	var paymentURL string
	switch method {
	case models.PaymentMethodZaloPay:
		paymentURL, err = s.createZaloPayOrder(payment, order)
	case models.PaymentMethodMoMo:
		paymentURL, err = s.createMoMoOrder(payment, order)
	case models.PaymentMethodVNPay:
		paymentURL, err = s.createVNPayOrder(payment, order)
	default:
		return nil, ErrInvalidPaymentMethod
	}

	if err != nil {
		return nil, err
	}

	payment.PaymentURL = paymentURL

	// Save payment
	if err := s.db.Create(payment).Error; err != nil {
		return nil, err
	}

	// Update order payment method
	order.PaymentMethod = method
	if err := s.orderRepo.Update(order); err != nil {
		return nil, err
	}

	return &dto.CreatePaymentResponse{
		OrderCode:  orderCode,
		PaymentURL: paymentURL,
		Method:     string(method),
	}, nil
}

func (s *paymentService) GetPaymentByOrderCode(orderCode string) (*models.Payment, error) {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return nil, ErrOrderNotFound
	}

	var payment models.Payment
	if err := s.db.Where("order_id = ?", order.ID).Order("created_at DESC").First(&payment).Error; err != nil {
		return nil, ErrPaymentNotFound
	}

	return &payment, nil
}

// ZaloPay Integration
func (s *paymentService) createZaloPayOrder(payment *models.Payment, order *models.Order) (string, error) {
	cfg := s.paymentCfg.ZaloPay

	appTime := time.Now().UnixMilli()
	appTransID := fmt.Sprintf("%s_%d", time.Now().Format("060102"), appTime)

	// Build request data
	data := map[string]interface{}{
		"app_id":       cfg.AppID,
		"app_user":     fmt.Sprintf("user_%d", order.UserID),
		"app_time":     appTime,
		"amount":       int64(payment.Amount),
		"app_trans_id": appTransID,
		"embed_data":   fmt.Sprintf(`{"order_code":"%s"}`, order.OrderCode),
		"item":         "[]",
		"description":  fmt.Sprintf("Payment for order %s", order.OrderCode),
		"bank_code":    "",
		"callback_url": cfg.CallbackURL,
	}

	// Generate MAC
	macData := fmt.Sprintf("%s|%d|%s|%s|%d|%s|%s",
		cfg.AppID, appTime, appTransID, fmt.Sprintf("user_%d", order.UserID),
		int64(payment.Amount), "[]", fmt.Sprintf(`{"order_code":"%s"}`, order.OrderCode))

	mac := s.hmacSHA256(macData, cfg.Key1)
	data["mac"] = mac

	// Send request
	jsonData, _ := json.Marshal(data)
	resp, err := http.Post(cfg.Endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", ErrPaymentFailed
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", ErrPaymentFailed
	}

	if returnCode, ok := result["return_code"].(float64); ok && returnCode == 1 {
		if orderURL, ok := result["order_url"].(string); ok {
			return orderURL, nil
		}
	}

	return "", ErrPaymentFailed
}

func (s *paymentService) HandleZaloPayCallback(data map[string]interface{}) error {
	cfg := s.paymentCfg.ZaloPay

	// Verify MAC
	dataStr, _ := data["data"].(string)
	mac, _ := data["mac"].(string)

	expectedMac := s.hmacSHA256(dataStr, cfg.Key2)
	if mac != expectedMac {
		return errors.New("invalid MAC")
	}

	// Parse callback data
	var callbackData map[string]interface{}
	if err := json.Unmarshal([]byte(dataStr), &callbackData); err != nil {
		return err
	}

	embedData, _ := callbackData["embed_data"].(string)
	var embed map[string]string
	json.Unmarshal([]byte(embedData), &embed)

	orderCode := embed["order_code"]
	zpTransID := fmt.Sprintf("%v", callbackData["zp_trans_id"])

	return s.updatePaymentStatus(orderCode, zpTransID, models.PaymentStatusSuccess)
}

// MoMo Integration
func (s *paymentService) createMoMoOrder(payment *models.Payment, order *models.Order) (string, error) {
	cfg := s.paymentCfg.MoMo

	requestID := uuid.New().String()
	orderID := payment.TransactionID

	rawSignature := fmt.Sprintf("accessKey=%s&amount=%d&extraData=&ipnUrl=%s&orderId=%s&orderInfo=%s&partnerCode=%s&redirectUrl=%s&requestId=%s&requestType=payWithMethod",
		cfg.AccessKey,
		int64(payment.Amount),
		cfg.CallbackURL,
		orderID,
		fmt.Sprintf("Payment for order %s", order.OrderCode),
		cfg.PartnerCode,
		cfg.ReturnURL,
		requestID,
	)

	signature := s.hmacSHA256(rawSignature, cfg.SecretKey)

	data := map[string]interface{}{
		"partnerCode": cfg.PartnerCode,
		"accessKey":   cfg.AccessKey,
		"requestId":   requestID,
		"amount":      int64(payment.Amount),
		"orderId":     orderID,
		"orderInfo":   fmt.Sprintf("Payment for order %s", order.OrderCode),
		"redirectUrl": cfg.ReturnURL,
		"ipnUrl":      cfg.CallbackURL,
		"extraData":   "",
		"requestType": "payWithMethod",
		"signature":   signature,
		"lang":        "en",
	}

	jsonData, _ := json.Marshal(data)
	resp, err := http.Post(cfg.Endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", ErrPaymentFailed
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", ErrPaymentFailed
	}

	if resultCode, ok := result["resultCode"].(float64); ok && resultCode == 0 {
		if payURL, ok := result["payUrl"].(string); ok {
			return payURL, nil
		}
	}

	return "", ErrPaymentFailed
}

func (s *paymentService) HandleMoMoCallback(data map[string]interface{}) error {
	resultCode, _ := data["resultCode"].(float64)
	orderID, _ := data["orderId"].(string)
	transID := fmt.Sprintf("%v", data["transId"])

	// Find payment by transaction ID
	var payment models.Payment
	if err := s.db.Where("transaction_id = ?", orderID).First(&payment).Error; err != nil {
		return err
	}

	// Get order
	var order models.Order
	if err := s.db.First(&order, payment.OrderID).Error; err != nil {
		return err
	}

	status := models.PaymentStatusFailed
	if resultCode == 0 {
		status = models.PaymentStatusSuccess
	}

	return s.updatePaymentStatusByPayment(&payment, &order, transID, status)
}

// VNPay Integration
func (s *paymentService) createVNPayOrder(payment *models.Payment, order *models.Order) (string, error) {
	cfg := s.paymentCfg.VNPay

	vnpParams := url.Values{}
	vnpParams.Set("vnp_Version", "2.1.0")
	vnpParams.Set("vnp_Command", "pay")
	vnpParams.Set("vnp_TmnCode", cfg.TmnCode)
	vnpParams.Set("vnp_Amount", fmt.Sprintf("%d", int64(payment.Amount)*100)) // VNPay uses VND * 100
	vnpParams.Set("vnp_CurrCode", "VND")
	vnpParams.Set("vnp_TxnRef", payment.TransactionID)
	vnpParams.Set("vnp_OrderInfo", fmt.Sprintf("Thanh toan don hang %s", order.OrderCode))
	vnpParams.Set("vnp_OrderType", "other")
	vnpParams.Set("vnp_Locale", "vn")
	vnpParams.Set("vnp_ReturnUrl", cfg.ReturnURL)
	vnpParams.Set("vnp_IpAddr", "127.0.0.1")
	vnpParams.Set("vnp_CreateDate", time.Now().Format("20060102150405"))

	// Sort params and create signature with URL-encoded values
	var keys []string
	for k := range vnpParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var signData strings.Builder
	for i, k := range keys {
		if i > 0 {
			signData.WriteString("&")
		}
		signData.WriteString(k)
		signData.WriteString("=")
		signData.WriteString(url.QueryEscape(vnpParams.Get(k)))
	}

	signature := s.hmacSHA512(signData.String(), cfg.HashSecret)
	vnpParams.Set("vnp_SecureHash", signature)

	return cfg.PaymentURL + "?" + vnpParams.Encode(), nil
}

func (s *paymentService) HandleVNPayCallback(params url.Values) error {
	cfg := s.paymentCfg.VNPay

	// Verify signature
	receivedHash := params.Get("vnp_SecureHash")
	params.Del("vnp_SecureHash")
	params.Del("vnp_SecureHashType")

	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var signData strings.Builder
	for i, k := range keys {
		if i > 0 {
			signData.WriteString("&")
		}
		signData.WriteString(k)
		signData.WriteString("=")
		signData.WriteString(params.Get(k))
	}

	expectedHash := s.hmacSHA512(signData.String(), cfg.HashSecret)
	if receivedHash != expectedHash {
		return errors.New("invalid signature")
	}

	txnRef := params.Get("vnp_TxnRef")
	responseCode := params.Get("vnp_ResponseCode")
	transactionNo := params.Get("vnp_TransactionNo")

	// Find payment
	var payment models.Payment
	if err := s.db.Where("transaction_id = ?", txnRef).First(&payment).Error; err != nil {
		return err
	}

	var order models.Order
	if err := s.db.First(&order, payment.OrderID).Error; err != nil {
		return err
	}

	status := models.PaymentStatusFailed
	if responseCode == "00" {
		status = models.PaymentStatusSuccess
	}

	return s.updatePaymentStatusByPayment(&payment, &order, transactionNo, status)
}

func (s *paymentService) updatePaymentStatus(orderCode string, providerTransID string, status models.PaymentStatus) error {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return err
	}

	var payment models.Payment
	if err := s.db.Where("order_id = ?", order.ID).Order("created_at DESC").First(&payment).Error; err != nil {
		return err
	}

	return s.updatePaymentStatusByPayment(&payment, order, providerTransID, status)
}

func (s *paymentService) updatePaymentStatusByPayment(payment *models.Payment, order *models.Order, providerTransID string, status models.PaymentStatus) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Update payment
		payment.Status = status
		payment.ProviderTransID = providerTransID
		if status == models.PaymentStatusSuccess {
			now := time.Now()
			payment.PaidAt = &now
		}

		if err := tx.Save(payment).Error; err != nil {
			return err
		}

		// Update order payment status
		order.PaymentStatus = status
		if status == models.PaymentStatusSuccess {
			order.Status = models.OrderStatusConfirmed
		}

		return tx.Save(order).Error
	})
}

func (s *paymentService) hmacSHA256(data, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *paymentService) hmacSHA512(data, key string) string {
	h := hmac.New(sha512.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// Helper to convert string to int
func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
