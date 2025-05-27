import grpc
from concurrent import futures
from gen.python.user_pb2_grpc import UserServiceServicer, add_UserServiceServicer_to_server
import logging

logging.basicConfig(level=logging.INFO)

class UserService(UserServiceServicer):
    pass

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    add_UserServiceServicer_to_server(UserService(), server)
    server.add_insecure_port('[::]:50052')
    logging.info("User Service running on :50052")
    server.start()
    server.wait_for_termination()

if __name__ == "__main__":
    serve()