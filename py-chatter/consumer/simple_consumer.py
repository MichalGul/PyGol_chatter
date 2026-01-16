import pika

# morph into connection manager
credentials = pika.PlainCredentials('guest', 'guest')
parameters = pika.ConnectionParameters('localhost', 5672, '/', credentials)
connection = pika.BlockingConnection(parameters)
channel = connection.channel()
channel.queue_declare(queue='test_queue', durable=True)


def callback(ch, method, properties, body):
    print(f" [x] Received {body.decode()}")
    reply = f"Pong from Python to: {body.decode()}"
    ch.basic_publish(exchange='', routing_key='pygol_queue', body=reply.encode(),
                     properties=pika.BasicProperties(delivery_mode=pika.DeliveryMode.Persistent))
    ch.basic_ack(delivery_tag=method.delivery_tag)

channel.basic_qos(prefetch_count=1)
channel.basic_consume(queue='test_queue', on_message_callback=callback)
print(' [*] Waiting for messages...')
channel.start_consuming()
