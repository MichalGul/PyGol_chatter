import pika
import uuid
import time
import json
# create class from this Require interface for message parsing and connection maintaining

MESSAGE_SCHEMA = {
  "conversation_id": "test123",
  "turn": 1,
  "max_turns": 10,
  "sender": "python",
  "message": "Hello from Python"
}

def prepare_connection():
    credentials = pika.PlainCredentials('guest', 'guest') # todo move to ENV
    parameters = pika.ConnectionParameters('localhost', 5672, '/', credentials)
    connection = pika.BlockingConnection(parameters=parameters)
    channel = connection.channel()
    channel.exchange_declare("llm.dialog.exchange", exchange_type="direct", durable=True)
    
    result = channel.queue_declare(queue="q.py.to.gol", durable=True)
    channel.queue_bind(exchange="llm.dialog.exchange", queue=result.method.queue)

    return channel, connection


# probably 2 queues one for send and 1 for recieve from golang
def send_message(channel, message_number=0):
    
    message = MESSAGE_SCHEMA.copy()
    message["turn"] = message_number
    message["message"] = f"Hello from Python! This is message number {message_number}"

    message_json = json.dumps(message)

    channel.basic_publish(exchange='llm.dialog.exchange',
                            routing_key='q.py.to.gol',
                            body=message_json,
                            properties=pika.BasicProperties(
                                delivery_mode=pika.DeliveryMode.Persistent,
                                content_type='application/json'
                            )
                          )


def send_messages_every(rabbit_connection: pika.BlockingConnection, rabbit_channel, timeout:float = 1.0, limit = 10):
    messages_send = 0
    while messages_send < limit:
        send_message(rabbit_channel, messages_send)
        messages_send+=1
        time.sleep(timeout)

    rabbit_connection.close()


def send_single_message(rabbit_channel, message: str, message_number=0):
    message_json = json.dumps({
        "conversation_id": "test123",
        "turn": message_number,
        "max_turns": 10,
        "sender": "python",
        "message": message
    })

    rabbit_channel.basic_publish(exchange='llm.dialog.exchange',
                            routing_key='q.py.to.gol',
                            body=message_json,
                            properties=pika.BasicProperties(
                                delivery_mode=pika.DeliveryMode.Persistent,
                                content_type='application/json'
                            )
                          )