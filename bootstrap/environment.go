package bootstrap

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Environment struct {
	DataBase   DataBase      `mapstructure:",squash"`
	Redis      Redis         `mapstructure:",squash"`
	HttpServer HttpServer    `mapstructure:",squash"`
	JWT        JWTConfig     `mapstructure:",squash"`
	OTP        OTPConfig     `mapstructure:",squash"`
	Seed       SeedConfig    `mapstructure:",squash"`
	Captcha    CaptchaConfig `mapstructure:",squash"`
	Email      EmailConfig   `mapstructure:",squash"`

	Invite     InvitationConfig `mapstructure:",squash"`
	Friendship FriendshipConfig `mapstructure:",squash"`
	RustFS     RustFSConfig     `mapstructure:",squash"`
	PoW        PoWConfig        `mapstructure:",squash"`
}

type FriendshipConfig struct {
	RequestCooldown time.Duration `mapstructure:"FRIENDSHIP_REQUEST_COOLDOWN"`
}

// PoWConfig holds settings for the two-layer rate limiter and stateless
// Proof-of-Work challenge system.
type PoWConfig struct {
	// HMACSecret is the signing key used to authenticate PoW challenges.
	// Must be kept secret; rotate it to invalidate all outstanding challenges.
	HMACSecret string `mapstructure:"POW_HMAC_SECRET" json:"-"`

	// ChallengeTTL is the maximum age of a challenge the server will accept.
	// Challenges older than this are rejected regardless of correctness.
	ChallengeTTL time.Duration `mapstructure:"POW_CHALLENGE_TTL"`

	// BlockTTL is how long a client remains blocked after hitting the L2
	// rate limit, unless they successfully solve a PoW challenge earlier.
	BlockTTL time.Duration `mapstructure:"POW_BLOCK_TTL"`

	// Difficulty is the number of leading zero bits required in a valid
	// solution hash.  Higher values mean more CPU work for the client.
	Difficulty int `mapstructure:"POW_DIFFICULTY"`

	// L1Limit is the outer (permissive) rate limit in requests per second.
	// Clients exceeding this receive a plain 429 with no challenge.
	L1Limit int `mapstructure:"POW_L1_LIMIT"`

	// L2Limit is the inner (strict) rate limit in requests per second.
	// L2Limit must be lower than L1Limit.
	// Clients exceeding this receive a 423 with a PoW challenge.
	L2Limit int `mapstructure:"POW_L2_LIMIT"`
}

type RustFSConfig struct {
	Endpoint        string `mapstructure:"RUSTFS_ENDPOINT"`
	AccessKeyID     string `mapstructure:"RUSTFS_ACCESS_KEY_ID"`
	SecretAccessKey string `mapstructure:"RUSTFS_SECRET_ACCESS_KEY" json:"-"`
	BucketName      string `mapstructure:"RUSTFS_BUCKET_NAME"`
	UseSSL          bool   `mapstructure:"RUSTFS_USE_SSL"`
	Region          string `mapstructure:"RUSTFS_REGION"`
	PublicURLPrefix string `mapstructure:"RUSTFS_PUBLIC_URL_PREFIX"`
	MaxConcurrency  int    `mapstructure:"RUSTFS_MAX_CONCURRENCY"`
}

type EmailConfig struct {
	DoratouEmail            string `mapstructure:"DORATOU_EMAIL"`
	DoratouEmailAppPassword string `mapstructure:"DORATOU_EMAIL_APP_PASSWORD" json:"-"`
	SMTPHost                string `mapstructure:"SMTP_HOST"`
	SMTPPort                string `mapstructure:"SMTP_HOST_PORT"`
	DeliveryEnabled         bool   `mapstructure:"EMAIL_DELIVERY_ENABLED"`
	DeliveryMode            string `mapstructure:"EMAIL_DELIVERY_MODE"`
	FontBaseURL             string `mapstructure:"EMAIL_FONT_BASE_URL"`
}

type CaptchaConfig struct {
	Enabled bool          `mapstructure:"CAPTCHA_ENABLED"`
	TTL     time.Duration `mapstructure:"CAPTCHA_TTL"`
}

type SeedConfig struct {
	UserName     string `mapstructure:"SEED_USER_NAME"`
	UserEmail    string `mapstructure:"SEED_USER_EMAIL"`
	UserPhone    string `mapstructure:"SEED_USER_PHONE"`
	UserPassword string `mapstructure:"SEED_USER_PASSWORD"`

	AdminName     string `mapstructure:"SEED_ADMIN_NAME"`
	AdminEmail    string `mapstructure:"SEED_ADMIN_EMAIL"`
	AdminPhone    string `mapstructure:"SEED_ADMIN_PHONE"`
	AdminPassword string `mapstructure:"SEED_ADMIN_PASSWORD"`
}

type JWTConfig struct {
	SecretKey       string        `mapstructure:"JWT_SECRET_KEY"`
	AccessTokenTTL  time.Duration `mapstructure:"ACCESS_TOKEN_TTL"`
	RefreshTokenTTL time.Duration `mapstructure:"REFRESH_TOKEN_TTL"`
}

type OTPConfig struct {
	TTL         time.Duration `mapstructure:"OTP_TTL"`
	VerifiedTTL time.Duration `mapstructure:"OTP_VERIFIED_TTL"`
	Digits      int           `mapstructure:"OTP_DIGITS"`
}

type DataBase struct {
	Host     string `mapstructure:"PDB_HOST"`
	Port     string `mapstructure:"PDB_PORT"`
	User     string `mapstructure:"PDB_USER"`
	Password string `mapstructure:"PDB_PASSWORD"`
	Name     string `mapstructure:"PDB_DB_NAME"`
	SSLMode  string `mapstructure:"PDB_SSL_MODE"`
}

type Redis struct {
	Host      string `mapstructure:"RDB_HOST"`
	Port      string `mapstructure:"RDB_PORT"`
	Password  string `mapstructure:"RDB_PASSWORD"`
	RDBNumber string `mapstructure:"RDB_DB_NUMBER"`
}

type HttpServer struct {
	Domain string `mapstructure:"HTTP_DOMAIN"`
	Port   string `mapstructure:"HTTP_PORT"`
}

type InvitationConfig struct {
	SecretKey                  string        `mapstructure:"INVITE_ENCRYPTION_SECRET"`
	DirectInviteRejectCooldown time.Duration `mapstructure:"GROUP_DIRECT_INVITE_REJECT_COOLDOWN"`
}

func NewEnvironment() (*Environment, error) {
	v := viper.New()
	v.SetConfigFile(".env")
	v.SetConfigType("env")

	for _, key := range EnvKeys {
		if err := v.BindEnv(key); err != nil {
			return nil, err
		}
	}
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read .env: %w", err)
	}
	var environment Environment
	if err := v.Unmarshal(&environment); err != nil {
		return nil, fmt.Errorf("read environment settings: %w", err)
	}

	return &environment, nil
}
