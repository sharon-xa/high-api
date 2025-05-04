-- +goose Up
INSERT INTO categories (name) VALUES
    ('Technology'),
    ('Science'),
    ('Health'),
    ('Travel'),
    ('Food'),
    ('Education'),
    ('Business'),
    ('Lifestyle'),
    ('Entertainment'),
    ('Sports');

-- +goose Down
DELETE FROM categories WHERE name IN (
    'Technology',
    'Science',
    'Health',
    'Travel',
    'Food',
    'Education',
    'Business',
    'Lifestyle',
    'Entertainment',
    'Sports'
);
