create table user_balance
(
    client_id  uuid,
    balance    bigint,
    revision   bigserial,
    created_at timestamp with time zone default now(),
    updated_at timestamp with time zone default now()
);

comment on column user_balance.client_id is 'id пользака';

comment on column user_balance.balance is 'кол-во денег на балансе';

comment on column user_balance.revision is 'ревизия баланса';

comment on column user_balance.created_at is 'время первой инициализации баланса, по сути отражает когда пользак был проинициализирован в выпуски и доехал до ядра процессинга';

comment on column user_balance.updated_at is 'время обновления баланса';

