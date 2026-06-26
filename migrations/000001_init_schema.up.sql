CREATE TABLE subscriptions(
    subscription_id UUID PRIMARY KEY,
    service_name TEXT NOT NULL,
    price INT NOT NULL CHECK (price > 0),
    user_id UUID UNIQUE NOT NULL,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP,
    CHECK (end_date IS NULL OR end_date >= start_date)
);

CREATE INDEX idx_subscriptions_user_service ON subscriptions (user_id, service_name);
