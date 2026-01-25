import pika
import sys
import os
import json

sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '..', '..')))

from py_chatter.utils import ExchangePyGol, QueueGolToPy

def prepare_connection():
    credentials = pika.PlainCredentials('guest', 'guest') # todo move to ENV
    parameters = pika.ConnectionParameters('localhost', 5672, '/', credentials)
    connection = pika.BlockingConnection(parameters=parameters)
    channel = connection.channel()
    
    result = channel.queue_declare(queue=QueueGolToPy, durable=True)
    channel.queue_bind(exchange=ExchangePyGol, queue=result.method.queue)

    return channel, connection


def callback(ch, method, properties, body):
    try:
        message_dict = json.loads(body.decode())
        print(f" [x] Received message: {message_dict}")

        # Access specific fields in the dictionary
        conversation_id = message_dict.get("conversation_id")
        turn = message_dict.get("turn")
        message = message_dict.get("message")

        print(f"Conversation ID: {conversation_id}, Turn: {turn}, Message: {message}")

        # todo handle message properly, probably pass message to orchestrator or LLM handler
        # or parse to some queue where llm will get and send response to producer
        # here we would send reply back to golang to queue q.gol.to.py

        # Orchestrator should handle replies and end of conversation logic
        # reply = f"Pong from Python to: {body.decode()}"
        # ch.basic_publish(exchange='', routing_key='pygol_queue', body=reply.encode(),
        #                  properties=pika.BasicProperties(delivery_mode=pika.DeliveryMode.Persistent))
        
        # when turns are over we should send some end signal or message


    except json.JSONDecodeError as e:
        print(f" [!] Error decoding JSON message: {e}")
        return
    
    finally:
        ch.basic_ack(delivery_tag=method.delivery_tag)
