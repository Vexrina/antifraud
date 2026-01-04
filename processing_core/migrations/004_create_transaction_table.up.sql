create table transactions_history
(
    id               INTEGER PRIMARY KEY,
    transaction_id   uuid UNIQUE,
    created_at       timestamp with time zone default now(),
    amount           bigint,
    currency         integer,
    merchant         text,
    country          text,
    sender_id        uuid,
    receiver_id      uuid,
    receiver_bic     text,
    atm_id           uuid,
    transaction_type transaction_type,
    revision         bigserial
);

comment on column transactions_history.id is 'уникальный ключ';

comment on column transactions_history.transaction_id is 'UUID транзакции из внешней системы';

comment on column transactions_history.created_at is 'дата и время операции как пришло к нам и обработалось у нас';

comment on column transactions_history.amount is 'сумма транзакции';

comment on column transactions_history.currency is 'валюта';

comment on column transactions_history.merchant is 'название/идентификатор магазина/сервиса';

comment on column transactions_history.country is 'ISO-код страны (RU/US/CA)';

comment on column transactions_history.sender_id is 'uuid отправителя';

comment on column transactions_history.receiver_id is 'uuid получателя';

comment on column transactions_history.receiver_bic is 'bic банка получателя';

comment on column transactions_history.atm_id is 'uuid банкомата';

comment on column transactions_history.transaction_type is 'тип транзакции';

comment on column transactions_history.revision is 'ревизия';