-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS identity.avatar_uploads (
    upload_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    
    filename VARCHAR(255) NOT NULL,
    size_bytes BIGINT NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    
    storage_key VARCHAR(500) NOT NULL,
    
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT avatar_uploads_status_check CHECK (status IN ('pending', 'completed', 'failed'))
);

CREATE INDEX idx_avatar_uploads_user_id ON identity.avatar_uploads(user_id);
CREATE INDEX idx_avatar_uploads_status ON identity.avatar_uploads(status);
CREATE INDEX idx_avatar_uploads_created_at ON identity.avatar_uploads(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS identity.avatar_uploads;
-- +goose StatementEnd
