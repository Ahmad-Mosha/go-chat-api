

CREATE INDEX idx_room_members_user_id ON room_members(user_id);
CREATE INDEX idx_messages_room_id ON messages(room_id);
CREATE INDEX idx_rooms_last_message_at ON rooms(last_message_at DESC);
