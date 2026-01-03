create table user_balance
(
    client_id uuid,
    balance   integer
);

comment on column user_balance.client_id is 'id пользака';

comment on column user_balance.balance is 'кол-во денег на балансе';

