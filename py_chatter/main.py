from producer.simple_producer import send_messages_every, prepare_connection as prepare_producer_connection
import threading
from consumer.simple_consumer import prepare_connection as prepare_consumer_connection, callback as consumer_callback
from py_chatter.utils import QueueGolToPy


def start_consumer(channel, connection):
    channel.basic_qos(prefetch_count=1)
    channel.basic_consume(queue=QueueGolToPy, on_message_callback=consumer_callback)
    print(' [*] Waiting for messages...')
    channel.start_consuming()



if __name__ == "__main__":
    print("Welcome to py-chatter")

    # start producer in separate thread
    channel, connection = prepare_producer_connection()
    producer_thread = threading.Thread(target=send_messages_every, args=(connection, channel, 2, 10))
    producer_thread.start()

    # start consumer in other thread
    channel_c, connection_c = prepare_consumer_connection()
    
    consumer_thread = threading.Thread(target=start_consumer, args=(channel_c, connection_c))
    consumer_thread.start()

    producer_thread.join()      
    consumer_thread.join()

