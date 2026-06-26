CREATE TABLE subscriptions(
    subscription_id UUID PRIMARY KEY,
    service_name TEXT NOT NULL,
    price INT NOT NULL CHECK (price > 0),
    user_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    CHECK (end_date IS NULL OR end_date >= start_date)
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions (user_id);
CREATE INDEX idx_subscriptions_user_service ON subscriptions (user_id, service_name);
