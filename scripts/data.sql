-- stock
INSERT INTO block(name, block_type, rarity, price, color_hex, stock)
VALUES('Red Ellie', '1x1+2x1', 'COMMON', 0, '#ff0000', true);

INSERT INTO block(name, block_type, rarity, price, color_hex, stock)
VALUES('Reverse Cyan Ellie', '2x1+1x1', 0, 'COMMON', '#00ffff', true);

INSERT INTO block(name, block_type, rarity, price, color_hex, stock)
VALUES('Gray Tank', '3x1+2x1', 'COMMON', 0, '#808080', true);

INSERT INTO block(name, block_type, rarity, price, color_hex, stock)
VALUES('Purple Firefly', '1x1', 'COMMON', 0, '#a020f0', true);

INSERT INTO block(name, block_type, rarity, price, color_hex, stock)
VALUES('Green Heavy Lifter', '2x2', 'COMMON', 0, '#90ee90', true);

-- buyable
INSERT INTO block(name, block_type, rarity, price, color_hex, stock)
VALUES('Silent Heavy Lifter', '2x2', 'LEGENDARY', 10, '#000080', false);

INSERT INTO block(name, block_type, rarity, price, color_hex, stock)
VALUES('Blue Heavy Lifter', '2x2', 'EPIC', 5, '#0000ff', false);

INSERT INTO block(name, block_type, rarity, price, color_hex, stock)
VALUES('Flashy Magenta Firefly', '1x1', 'EPIC', 5, '#FF00FF', false);

INSERT INTO block(name, block_type, rarity, price, color_hex, stock)
VALUES('Flashy Magenta Firefly', '1x1', 'LEGENDARY', 10, '#FF00FF', false);

INSERT INTO block(name, block_type, rarity, price, color_hex, stock)
VALUES('Yellow pipe', '4x1', 'RARE', 3, '#FFFF00', false);