-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
	id bigserial NOT NULL,
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	"name" text NULL,
	hash text NULL,
	CONSTRAINT users_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_users_deleted_at ON public.users USING btree (deleted_at);

CREATE TABLE balances (
	id bigserial NOT NULL,
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	user_id int8 NULL,
	balance numeric(10,2) NULL,
	withdrawn numeric(10,2) NULL,
	CONSTRAINT balances_pkey PRIMARY KEY (id),
	CONSTRAINT fk_balances_user FOREIGN KEY (user_id) REFERENCES public.users(id)
);
CREATE INDEX idx_balances_deleted_at ON public.balances (deleted_at timestamptz_ops);

CREATE TABLE operations (
	id bigserial NOT NULL,
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	user_id int8 NULL,
	"order" text NULL,
	value numeric(10,2) NULL,
	operation_type text NULL,
	CONSTRAINT operations_pkey PRIMARY KEY (id),
	CONSTRAINT fk_users_operation FOREIGN KEY (user_id) REFERENCES public.users(id)
);
CREATE INDEX idx_operations_deleted_at ON public.operations (deleted_at timestamptz_ops);

CREATE TABLE orders (
	id bigserial NOT NULL,
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	deleted_at timestamptz NULL,
	user_id int8 NULL,
	"name" text NULL,
	hash text NULL,
	"number" text NULL,
	accrual numeric(10,2) NULL,
	status text NULL,
	CONSTRAINT orders_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_orders_deleted_at ON public.orders USING btree (deleted_at, deleted_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE balances;
DROP TABLE operations;
DROP TABLE orders;
DROP TABLE users;
-- +goose StatementEnd