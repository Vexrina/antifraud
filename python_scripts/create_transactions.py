import grpc
import uuid
import random
from datetime import datetime
import time
from google.protobuf.timestamp_pb2 import Timestamp

import core_pb2
import core_pb2_grpc

# Конфигурация
NUM_UUIDS = 100_000

SERVER = "localhost:9090"  # порт gRPC сервера
N_SECONDS = 0.01

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

print(f"start generate {NUM_UUIDS} for test")
uuids = []
for i in range(NUM_UUIDS):
    uuids.append(str(uuid.uuid4()))
    
print(f"end generate {NUM_UUIDS} for test")
currentID = 631794 + 100

def random_transaction(user_id: uuid.UUID|None) -> core_pb2.Transaction:
    global currentID
    ts = Timestamp()
    ts.FromDatetime(datetime.utcnow())
    currentID += 1
    return core_pb2.Transaction(
        id=currentID,
        transaction_id=str(uuid.uuid4()),
        created_at=ts,
        amount=random.randint(1_000, 200_000)*100, # в копейках!
        currency=random.choice(CURRENCIES),
        merchant=random.choice(MERCHANTS),
        country=random.choice(COUNTRIES),
        sender_id = str(user_id) if user_id is not None else str(uuids[random.randint(0, NUM_UUIDS - 1)])
    )

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

def main():
    i = 400_000
    with grpc.insecure_channel(SERVER) as channel:
        stub = core_pb2_grpc.CoreStub(channel)
        print("*" * 20)
        print("Start add balance for clients")
        print("*" * 20)
        now = datetime.now()
        j = 0
        for uuid in uuids:
            method = 'CashIn'
            req = random_request(method, uuid)
            try:
                resp = stub.CashIn(req)
                if j % 1000 == 0:
                    print(f"REQ# {j+1}")
                j+=1
            except Exception as e:
                continue
        end = datetime.now()
        print("*" * 20)
        print(f"End add balance for clients, time taken: {end - now}")
        print("*" * 20)
        print("Start create transactions")
        print("*" * 20)
        now = datetime.now()
        while True:
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
                if i % 1000 == 0:
                    print(f"REQ# {i+1}")
            except Exception as e:
                print(f"Ignored error in {method}: {e}")
            # time.sleep(N_SECONDS)
            i+=1
        end = datetime.now()
        print("*" * 20)
        print(f"End create transactions, time taken: {end - now}")
        print("*" * 20)

if __name__ == "__main__":
    main()
