package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Google   GoogleConfig
	Payment  PaymentConfig
	Minio    MinioConfig
}

type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type GoogleConfig struct {
	ClientID     string
	ClientSecret string
}

type PaymentConfig struct {
	ZaloPay ZaloPayConfig
	MoMo    MoMoConfig
	VNPay   VNPayConfig
}

type ZaloPayConfig struct {
	AppID       string
	Key1        string
	Key2        string
	Endpoint    string
	CallbackURL string
}

type MoMoConfig struct {
	PartnerCode string
	AccessKey   string
	SecretKey   string
	Endpoint    string
	CallbackURL string
	ReturnURL   string
}

type VNPayConfig struct {
	TmnCode     string
	HashSecret  string
	PaymentURL  string
	ReturnURL   string
}

type ServerConfig struct {
	Port         string
	Mode         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Service  string
	SSLMode  string
}

type JWTConfig struct {
	Secret          string
	ExpirationHours int
	Issuer          string
}

func Load() *Config {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("SERVER_MODE", "debug")
	viper.SetDefault("SERVER_READ_TIMEOUT", 10)
	viper.SetDefault("SERVER_WRITE_TIMEOUT", 10)

	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "1521")
	viper.SetDefault("DB_USER", "florus")
	viper.SetDefault("DB_PASSWORD", "florus123")
	viper.SetDefault("DB_SERVICE", "XEPDB1")
	viper.SetDefault("DB_SSL_MODE", "disable")

	viper.SetDefault("JWT_SECRET", "florus-secret-key-change-in-production")
	viper.SetDefault("JWT_EXPIRATION_HOURS", 24)
	viper.SetDefault("JWT_ISSUER", "florus-api")

	// Google OAuth defaults
	viper.SetDefault("GOOGLE_CLIENT_ID", "")
	viper.SetDefault("GOOGLE_CLIENT_SECRET", "")

	// ZaloPay defaults (sandbox)
	viper.SetDefault("ZALOPAY_APP_ID", "2553")
	viper.SetDefault("ZALOPAY_KEY1", "PcY4iZIKFCIdgZvA6ueMcMHHmU1y8R1W")
	viper.SetDefault("ZALOPAY_KEY2", "kLtgPl8HHhfvMuDHPwKfgfsY4AaYQCT0")
	viper.SetDefault("ZALOPAY_ENDPOINT", "https://sb-openapi.zalopay.vn/v2/create")
	viper.SetDefault("ZALOPAY_CALLBACK_URL", "http://localhost:8081/api/payments/callback/zalopay")

	// MoMo defaults (sandbox)
	viper.SetDefault("MOMO_PARTNER_CODE", "MOMO")
	viper.SetDefault("MOMO_ACCESS_KEY", "F8BBA842ECF85")
	viper.SetDefault("MOMO_SECRET_KEY", "K951B6PE1waDMi640xX08PD3vg6EkVlz")
	viper.SetDefault("MOMO_ENDPOINT", "https://test-payment.momo.vn/v2/gateway/api/create")
	viper.SetDefault("MOMO_CALLBACK_URL", "http://localhost:8081/api/payments/callback/momo")
	viper.SetDefault("MOMO_RETURN_URL", "http://localhost:3000/payment/result")

	// VNPay defaults (sandbox)
	viper.SetDefault("VNPAY_TMN_CODE", "DEMO")
	viper.SetDefault("VNPAY_HASH_SECRET", "DEMO_SECRET_KEY")
	viper.SetDefault("VNPAY_PAYMENT_URL", "https://sandbox.vnpayment.vn/paymentv2/vpcpay.html")
	viper.SetDefault("VNPAY_RETURN_URL", "http://localhost:3000/payment/result")

	// MinIO defaults
	viper.SetDefault("MINIO_ENDPOINT", "localhost:9000")
	viper.SetDefault("MINIO_ACCESS_KEY", "minioadmin")
	viper.SetDefault("MINIO_SECRET_KEY", "minioadmin123")
	viper.SetDefault("MINIO_BUCKET", "florus-images")
	viper.SetDefault("MINIO_USE_SSL", false)

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Config file not found, using defaults and environment variables: %v", err)
	}

	return &Config{
		Server: ServerConfig{
			Port:         viper.GetString("SERVER_PORT"),
			Mode:         viper.GetString("SERVER_MODE"),
			ReadTimeout:  time.Duration(viper.GetInt("SERVER_READ_TIMEOUT")) * time.Second,
			WriteTimeout: time.Duration(viper.GetInt("SERVER_WRITE_TIMEOUT")) * time.Second,
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			Service:  viper.GetString("DB_SERVICE"),
			SSLMode:  viper.GetString("DB_SSL_MODE"),
		},
		JWT: JWTConfig{
			Secret:          viper.GetString("JWT_SECRET"),
			ExpirationHours: viper.GetInt("JWT_EXPIRATION_HOURS"),
			Issuer:          viper.GetString("JWT_ISSUER"),
		},
		Google: GoogleConfig{
			ClientID:     viper.GetString("GOOGLE_CLIENT_ID"),
			ClientSecret: viper.GetString("GOOGLE_CLIENT_SECRET"),
		},
		Payment: PaymentConfig{
			ZaloPay: ZaloPayConfig{
				AppID:       viper.GetString("ZALOPAY_APP_ID"),
				Key1:        viper.GetString("ZALOPAY_KEY1"),
				Key2:        viper.GetString("ZALOPAY_KEY2"),
				Endpoint:    viper.GetString("ZALOPAY_ENDPOINT"),
				CallbackURL: viper.GetString("ZALOPAY_CALLBACK_URL"),
			},
			MoMo: MoMoConfig{
				PartnerCode: viper.GetString("MOMO_PARTNER_CODE"),
				AccessKey:   viper.GetString("MOMO_ACCESS_KEY"),
				SecretKey:   viper.GetString("MOMO_SECRET_KEY"),
				Endpoint:    viper.GetString("MOMO_ENDPOINT"),
				CallbackURL: viper.GetString("MOMO_CALLBACK_URL"),
				ReturnURL:   viper.GetString("MOMO_RETURN_URL"),
			},
			VNPay: VNPayConfig{
				TmnCode:    viper.GetString("VNPAY_TMN_CODE"),
				HashSecret: viper.GetString("VNPAY_HASH_SECRET"),
				PaymentURL: viper.GetString("VNPAY_PAYMENT_URL"),
				ReturnURL:  viper.GetString("VNPAY_RETURN_URL"),
			},
		},
		Minio: MinioConfig{
			Endpoint:  viper.GetString("MINIO_ENDPOINT"),
			AccessKey: viper.GetString("MINIO_ACCESS_KEY"),
			SecretKey: viper.GetString("MINIO_SECRET_KEY"),
			Bucket:    viper.GetString("MINIO_BUCKET"),
			UseSSL:    viper.GetBool("MINIO_USE_SSL"),
		},
	}
}
