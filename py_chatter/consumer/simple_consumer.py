import pika
import sys
import os

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
    print(f" [x] Received {body.decode()}")

    # todo process message here unmarshalling, calling LLM, etc.
    # here we would send reply back to golang to queue q.gol.to.py

    # Orchestrator should handle replies and end of conversation logic
    # reply = f"Pong from Python to: {body.decode()}"
    # ch.basic_publish(exchange='', routing_key='pygol_queue', body=reply.encode(),
    #                  properties=pika.BasicProperties(delivery_mode=pika.DeliveryMode.Persistent))
    ch.basic_ack(delivery_tag=method.delivery_tag)


if __name__ == "__main__":
    print("Starting Python consumer...")
    channel, connection = prepare_connection()


    channel.basic_qos(prefetch_count=1)
    channel.basic_consume(queue=QueueGolToPy, on_message_callback=callback)
    print(' [*] Waiting for messages...')
    channel.start_consuming()
