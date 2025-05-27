import grpc
from concurrent import futures
from gen.python.test_pb2_grpc import TestServiceServicer, add_TestServiceServicer_to_server
import logging

logging.basicConfig(level=logging.INFO)

class TestService(TestServiceServicer):
    pass

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    add_TestServiceServicer_to_server(TestService(), server)
    server.add_insecure_port('[::]:50053')
    logging.info("Test Service running on :50053")
    server.start()
    server.wait_for_termination()

if __name__ == "__main__":
    serve()