CREATE TYPE transaction_type AS ENUM (
  'internal',
  'cash_in',
  'cash_out',
  'sbp_outgoing'
);

comment on type transaction_type is 'тип транзакции';