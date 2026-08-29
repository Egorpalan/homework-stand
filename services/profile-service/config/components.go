package config

import "time"

type GrpcServer struct {
	Host              string        `yaml:"host"`
	Port              int           `yaml:"port" env:"GRPC_PORT"`
	MaxConnectionIdle time.Duration `yaml:"max_connection_idle"`
	MaxConnectionAge  time.Duration `yaml:"max_connection_age"`
	Timeout           time.Duration `yaml:"timeout"`
	Time              time.Duration `yaml:"time"`
}

type Postgres struct {
	Host               string        `yaml:"host"`
	Port               int32         `yaml:"port"`
	User               string        `yaml:"user"`
	Password           string        `yaml:"password"`
	Database           string        `yaml:"database"`
	MaxConnections     int32         `yaml:"max_connections"`
	MinConnections     int32         `yaml:"min_connections"`
	MaxIdleConnections int32         `yaml:"max_idle_connections"`
	MaxConnLifetime    time.Duration `yaml:"max_conn_lifetime"`
}

type Kafka struct {
	Brokers       []string `yaml:"brokers"`
	ConsumerGroup string   `yaml:"consumer_group"`
}

type RedisCluster struct {
	Addresses    []string      `yaml:"addresses"`
	MaxRetries   int           `yaml:"max_retries"`
	DialTimeout  time.Duration `yaml:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type Cache struct {
	KeyVersion       string        `yaml:"key_version"`
	SoftTTL          time.Duration `yaml:"soft_ttl"`
	TTL              time.Duration `yaml:"hard_ttl"`
	JitterPercentage float64       `yaml:"jitter_percentage"`
	Negative         struct {
		NotFoundTTL time.Duration `yaml:"not_found_ttl"`
	} `yaml:"negative"`
}

type HttpServer struct {
	Port      int `yaml:"port" env:"HTTP_PORT"`
	AdminPort int `yaml:"admin_port" env:"ADMIN_HTTP_PORT"`
}

type Graceful struct {
	Timeout time.Duration `yaml:"timeout"`
}

type Categories struct {
	FilePath string `yaml:"file_path"`
}
