-- +goose Up
INSERT INTO categories (name) VALUES
('Programming'),
('Software Architecture'),
('Tools & Workflows'),
('Databases'),
('Productivity'),
('Security'),
('Testing & Debugging'),
('Career & Growth'),
('Operating Systems'),
('Infrastructure & Cloud'),
('Web & Internet'),
('Open Source');

-- +goose Down
DELETE FROM categories
WHERE name IN (
    'Programming',
    'Software Architecture',
    'Tools & Workflows',
    'Databases',
    'Productivity',
    'Security',
    'Testing & Debugging',
    'Career & Growth',
    'Operating Systems',
    'Infrastructure & Cloud',
    'Web & Internet',
    'Open Source'
);
