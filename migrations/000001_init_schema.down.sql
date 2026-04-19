-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS room_members;
DROP TABLE IF EXISTS rooms;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
