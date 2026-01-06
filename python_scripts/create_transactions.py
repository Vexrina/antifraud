import grpc
import uuid
import random
from datetime import datetime
from google.protobuf.timestamp_pb2 import Timestamp

import core_pb2
import core_pb2_grpc

# Конфигурация
NUM_UUIDS = 1000

SERVER = "localhost:8081"  # порт gRPC сервера
N_SECONDS = 2

CURRENCIES = [1, 2, 3]
COUNTRIES = ["RU", "US", "CA"]
MERCHANTS = ["Amazon", "Ozon", "YandexMarket", "LocalShop"]
METHODS = ["Internal", "SbpOutgoing", "CashIn", "CashOut"]

# Список эндпоинтов и методов
ENDPOINTS = [
    "/v1/internal",
    "/v1/sbp-outgoing",
    "/v1/cash/in",
    "/v1/cash/out"
]

uuids = [str(uuid.uuid4()) for _ in range(NUM_UUIDS)]
currentID = 10

# Генератор случайной транзакции (JSON)
def random_transaction(user_id: uuid.UUID|None) -> core_pb2.Transaction:
    global currentID
    ts = Timestamp()
    ts.FromDatetime(datetime.utcnow())
    currentID += 1
    return core_pb2.Transaction(
        id=currentID,
        transaction_id=str(uuid.uuid4()),
        created_at=ts,
        amount=random.randint(100, 100_000),
        currency=random.choice(CURRENCIES),
        merchant=random.choice(MERCHANTS),
        country=random.choice(COUNTRIES),
        sender_id = str(user_id) if user_id is not None else str(uuids[random.randint(0, NUM_UUIDS - 1)])
    )

# Генератор случайного запроса под каждый эндпоинт
def random_request(method, clientID: uuid.UUID|None=None):
    tx = random_transaction(clientID)

    senderid = tx.sender_id

    req_id = str(uuid.uuid4())
    receiver_id = str(uuid.uuid4())
    atm_id = str(uuid.uuid4())
    bic = "BIC" + str(random.randint(1000, 9999))
    if method == "SbpOutgoing":
        tx.merchant = "SbpIntegration"
    if method == "CashIn" or method == "CashOut":
        tx.merchant = "ATMIntegration"

    if method == "Internal":
        return core_pb2.InternalRequest(id=req_id, transaction=tx, receiver_id=random.choice([u for u in uuids if u != senderid]))
    elif method == "SbpOutgoing":
        return core_pb2.SbpOutgoingRequest(id=req_id, transaction=tx, receiver_id=receiver_id, bic=bic)
    elif method == "CashIn":
        return core_pb2.CashInRequest(id=req_id, transaction=tx, atm_id=atm_id)
    elif method == "CashOut":
        return core_pb2.CashOutRequest(id=req_id, transaction=tx, atm_id=atm_id)

# Основной async loop
def main():
    i = 0
    with grpc.insecure_channel(SERVER) as channel:
        stub = core_pb2_grpc.CoreStub(channel)
        for uuid in uuids:
            method = 'CashIn'
            req = random_request(method, uuid)
            try:
                resp = stub.CashIn(req)
                print(f"{method} called, status: {resp.new_status}")
            except Exception as e:
                print(f"Ignored error in {method}: {e}")

        while i!=10_000:
            method = random.choice(METHODS)
            req = random_request(method)
            try:
                if method == "Internal":
                    resp = stub.Internal(req)
                elif method == "SbpOutgoing":
                    resp = stub.SbpOutgoing(req)
                elif method == "CashIn":
                    resp = stub.CashIn(req)
                elif method == "CashOut":
                    resp = stub.CashOut(req)
                print(f"{method} called, status: {resp.new_status}")
            except Exception as e:
                print(f"Ignored error in {method}: {e}")
            
            import time
            time.sleep(N_SECONDS)
            i+=1

if __name__ == "__main__":
    main()
