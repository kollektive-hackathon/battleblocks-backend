CREATE TABLE custodial_wallet
(
    id         BIGSERIAL PRIMARY KEY,
    resourceId TEXT         NOT NULL,
    public_key TEXT         NOT NULL,
    address    VARCHAR(255),

    UNIQUE (resourceId),
    UNIQUE (public_key),
    UNIQUE (address)
);

CREATE TABLE battleblocks_user
(
    id                          BIGSERIAL PRIMARY KEY,
    email                       VARCHAR(128) NOT NULL,
    username                    VARCHAR(32)  NOT NULL,
    custodial_wallet_id         BIGINT       NOT NULL,
    self_custody_wallet_address VARCHAR(255),
    google_identity_id          VARCHAR(128),

    UNIQUE (email),
    UNIQUE (username),
    UNIQUE (self_custody_wallet_address),
    UNIQUE (google_identity_id),

    CONSTRAINT fk_user_custodial_wallet_id FOREIGN KEY (custodial_wallet_id) REFERENCES custodial_wallet (id)
);

CREATE TYPE RARITY AS enum ('COMMON', 'RARE', 'EPIC', 'LEGENDARY');
CREATE TYPE BLOCK_TYPE AS enum ('a12b1', 'a1b12', 'a123b2', 'a1', 'a12b12', 'a1234');

CREATE TABLE block
(
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT       NOT NULL,
    block_type BLOCK_TYPE NOT NULL,
    rarity     RARITY     NOT NULL,
    price      BIGINT     NOT NULL,
    color_hex  CHAR(7),
    stock      BOOL       NOT NULL
);

CREATE TYPE GAME_STATUS AS enum ('CREATED', 'PREPARING', 'PLAYING', 'FINISHED');

CREATE TABLE game
(
    id                 BIGSERIAL PRIMARY KEY,
    flow_id            BIGINT,
    owner_id           BIGINT      NOT NULL,
    challenger_id      BIGINT,
    game_status        GAME_STATUS NOT NULL,
    stake              BIGINT      NOT NULL,
    time_started       BIGINT,
    time_created       BIGINT,
    winner             BIGINT
);

CREATE TABLE block_placement
(
    id          BIGINT PRIMARY KEY,
    user_id     BIGINT  NOT NULL,
    game_id     BIGINT  NOT NULL,
    block_id    BIGINT  NOT NULL,
    coordinateX INTEGER NOT NULL,
    coordinateY INTEGER NOT NULL,

    UNIQUE (game_id, coordinateX, coordinateY),

    CONSTRAINT fk_block_placement_user_id FOREIGN KEY (user_id) REFERENCES battleblocks_user (id),
    CONSTRAINT fk_block_placement_game_id FOREIGN KEY (game_id) REFERENCES game (id),
    CONSTRAINT fk_block_placement_block_id FOREIGN KEY (block_id) REFERENCES block (id)
);



CREATE TABLE nft
(
    id       BIGSERIAL NOT NULL,
    flow_id  BIGINT,
    block_id BIGINT    NOT NULL,

    UNIQUE (flow_id),

    CONSTRAINT fk_nft_block_id FOREIGN KEY (block_id) REFERENCES block (id)
);

CREATE TABLE move_history
(
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT  NOT NULL,
    game_id     BIGINT  NOT NULL,
    coordinateX INTEGER NOT NULL,
    coordinateY INTEGER NOT NULL,
    played_at   BIGINT,

    UNIQUE (game_id, coordinateX, coordinateY),

    CONSTRAINT fk_move_history_user_id FOREIGN KEY (user_id) REFERENCES battleblocks_user (id),
    CONSTRAINT fk_move_history_game_id FOREIGN KEY (game_id) REFERENCES game (id)
);

CREATE TABLE user_block_inventory
(
    user_id  BIGINT NOT NULL,
    block_id BIGINT NOT NULL,
    active   bool   NOT NULL,

    PRIMARY KEY (user_id, block_id),
    CONSTRAINT fk_user_block_inventory_user_id FOREIGN KEY (user_id) REFERENCES battleblocks_user (id),
    CONSTRAINT fk_user_block_inventory_block_id FOREIGN KEY (block_id) REFERENCES block (id)
);

CREATE TABLE nft_purchase_history
(
    nft_id       BIGINT                   NOT NULL PRIMARY KEY,
    buyer_id     BIGINT                   NOT NULL,
    purchased_at BIGINT NOT NULL,

    CONSTRAINT nft_purchase_history_buyer_id FOREIGN KEY (buyer_id) REFERENCES battleblocks_user (id)
);

CREATE TABLE game_grid_point (
    game_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    block_present BOOL NOT NULL,
    coordinate_x INTEGER NOT NULL,
    coordinate_y INTEGER NOT NULL,
    nonce text NOT NULL,

    PRIMARY KEY (game_id, user_id)
);