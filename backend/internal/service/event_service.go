package service

import (
	"encoding/json"
	"time"

	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/internal/repository"
)

type EventService interface {
	// Event logging
	LogEvent(event *models.UserEvent) error
	LogProductView(userID *uint, sessionID string, productID uint, categoryID *uint) error
	LogAddToCart(userID *uint, sessionID string, productID uint, quantity int, price float64) error
	LogRemoveFromCart(userID *uint, sessionID string, productID uint) error
	LogPurchase(userID *uint, sessionID string, productID uint, orderID uint, quantity int, price float64) error
	LogSearch(userID *uint, sessionID string, query string) error
	LogCategoryView(userID *uint, sessionID string, categoryID uint) error
	LogRating(userID uint, productID uint, rating int) error

	// Analytics
	GetUserBehaviorProfile(userID uint) (*models.UserBehaviorProfile, error)
	GetProductEventStats(productID uint) (*models.ProductEventStats, error)
	GetEventStats(days int) ([]models.UserEventStats, error)
	GetUserRecentProducts(userID uint, limit int) ([]uint, error)

	// Recommendations data
	GetPopularProducts(limit int, days int) ([]uint, error)
	GetFrequentlyBoughtTogether(productID uint, limit int) ([]uint, error)
	GetCoViewedProducts(productID uint, limit int) ([]uint, error)

	// Maintenance
	CleanupOldEvents(days int) (int64, error)
}

type eventService struct {
	eventRepo   repository.EventRepository
	kafkaClient KafkaClient // Optional Kafka client for streaming
}

// KafkaClient interface for event streaming (optional)
type KafkaClient interface {
	SendEvent(topic string, key string, value []byte) error
}

func NewEventService(eventRepo repository.EventRepository, kafkaClient KafkaClient) EventService {
	return &eventService{
		eventRepo:   eventRepo,
		kafkaClient: kafkaClient,
	}
}

// LogEvent logs a generic event
func (s *eventService) LogEvent(event *models.UserEvent) error {
	// Save to database
	if err := s.eventRepo.Create(event); err != nil {
		return err
	}

	// Stream to Kafka if available
	s.streamToKafka(event)

	return nil
}

// LogProductView logs a product view event
func (s *eventService) LogProductView(userID *uint, sessionID string, productID uint, categoryID *uint) error {
	event := &models.UserEvent{
		UserID:     userID,
		SessionID:  sessionID,
		EventType:  models.EventProductView,
		ProductID:  &productID,
		CategoryID: categoryID,
	}
	return s.LogEvent(event)
}

// LogAddToCart logs an add to cart event
func (s *eventService) LogAddToCart(userID *uint, sessionID string, productID uint, quantity int, price float64) error {
	event := &models.UserEvent{
		UserID:    userID,
		SessionID: sessionID,
		EventType: models.EventAddToCart,
		ProductID: &productID,
		Quantity:  quantity,
		Price:     price,
	}
	return s.LogEvent(event)
}

// LogRemoveFromCart logs a remove from cart event
func (s *eventService) LogRemoveFromCart(userID *uint, sessionID string, productID uint) error {
	event := &models.UserEvent{
		UserID:    userID,
		SessionID: sessionID,
		EventType: models.EventRemoveFromCart,
		ProductID: &productID,
	}
	return s.LogEvent(event)
}

// LogPurchase logs a purchase event
func (s *eventService) LogPurchase(userID *uint, sessionID string, productID uint, orderID uint, quantity int, price float64) error {
	event := &models.UserEvent{
		UserID:    userID,
		SessionID: sessionID,
		EventType: models.EventPurchase,
		ProductID: &productID,
		OrderID:   &orderID,
		Quantity:  quantity,
		Price:     price,
	}
	return s.LogEvent(event)
}

// LogSearch logs a search event
func (s *eventService) LogSearch(userID *uint, sessionID string, query string) error {
	event := &models.UserEvent{
		UserID:      userID,
		SessionID:   sessionID,
		EventType:   models.EventSearch,
		SearchQuery: query,
	}
	return s.LogEvent(event)
}

// LogCategoryView logs a category view event
func (s *eventService) LogCategoryView(userID *uint, sessionID string, categoryID uint) error {
	event := &models.UserEvent{
		UserID:     userID,
		SessionID:  sessionID,
		EventType:  models.EventCategoryView,
		CategoryID: &categoryID,
	}
	return s.LogEvent(event)
}

// LogRating logs a product rating event
func (s *eventService) LogRating(userID uint, productID uint, rating int) error {
	event := &models.UserEvent{
		UserID:    &userID,
		EventType: models.EventRating,
		ProductID: &productID,
		Rating:    &rating,
	}
	return s.LogEvent(event)
}

// GetUserBehaviorProfile returns user behavior profile for recommendations
func (s *eventService) GetUserBehaviorProfile(userID uint) (*models.UserBehaviorProfile, error) {
	return s.eventRepo.GetUserBehaviorProfile(userID)
}

// GetProductEventStats returns product event statistics
func (s *eventService) GetProductEventStats(productID uint) (*models.ProductEventStats, error) {
	return s.eventRepo.GetProductEventStats(productID)
}

// GetEventStats returns overall event statistics
func (s *eventService) GetEventStats(days int) ([]models.UserEventStats, error) {
	since := time.Now().AddDate(0, 0, -days)
	return s.eventRepo.GetEventStats(since)
}

// GetUserRecentProducts returns recent products viewed by user
func (s *eventService) GetUserRecentProducts(userID uint, limit int) ([]uint, error) {
	return s.eventRepo.GetUserProductViews(userID, limit)
}

// GetPopularProducts returns most popular products
func (s *eventService) GetPopularProducts(limit int, days int) ([]uint, error) {
	since := time.Now().AddDate(0, 0, -days)
	return s.eventRepo.GetPopularProducts(limit, since)
}

// GetFrequentlyBoughtTogether returns frequently bought together products
func (s *eventService) GetFrequentlyBoughtTogether(productID uint, limit int) ([]uint, error) {
	return s.eventRepo.GetFrequentlyBoughtTogether(productID, limit)
}

// GetCoViewedProducts returns co-viewed products
func (s *eventService) GetCoViewedProducts(productID uint, limit int) ([]uint, error) {
	return s.eventRepo.GetCoViewedProducts(productID, limit)
}

// CleanupOldEvents removes old events
func (s *eventService) CleanupOldEvents(days int) (int64, error) {
	before := time.Now().AddDate(0, 0, -days)
	return s.eventRepo.DeleteOldEvents(before)
}

// streamToKafka sends event to Kafka for real-time processing
func (s *eventService) streamToKafka(event *models.UserEvent) {
	if s.kafkaClient == nil {
		return
	}

	// Prepare Kafka message
	eventData, err := json.Marshal(map[string]interface{}{
		"event_id":     event.ID,
		"user_id":      event.UserID,
		"session_id":   event.SessionID,
		"event_type":   event.EventType,
		"product_id":   event.ProductID,
		"category_id":  event.CategoryID,
		"order_id":     event.OrderID,
		"search_query": event.SearchQuery,
		"quantity":     event.Quantity,
		"price":        event.Price,
		"rating":       event.Rating,
		"timestamp":    event.CreatedAt.Unix(),
	})

	if err != nil {
		return
	}

	// Send to Kafka topic
	topic := "florus-user-events"
	key := event.SessionID
	if event.UserID != nil {
		key = string(rune(*event.UserID))
	}

	go s.kafkaClient.SendEvent(topic, key, eventData)
}
