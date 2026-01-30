import sys
from producer.simple_producer import send_single_message, send_messages_every, prepare_connection as prepare_producer_connection
import threading
from consumer.simple_consumer import prepare_connection as prepare_consumer_connection, callback as consumer_callback
from py_chatter.utils import QueueGolToPy
import os

def start_consumer(channel, connection):
    channel.basic_qos(prefetch_count=1)
    channel.basic_consume(queue=QueueGolToPy, on_message_callback=consumer_callback)
    print(' [*] Waiting for messages...')
    channel.start_consuming()


def clear_screen():
    """Clear the console screen"""
    os.system('cls' if os.name == 'nt' else 'clear')

def main_menu():
    """Display the main interactive menu"""

    consumer_thread = None
    consumer_connection = None
    message_counter = 0
    clear_screen()
    while True:
        
        print("Welcome to py-chatter")
        print("1. Send Message")
        print("2. Start Consumer in Background")
        print("3. Exit")

        choice = input("\nEnter your choice: ")   

        if choice == '1':
            channel, connection = prepare_producer_connection()
            message = input("Enter your message: ")
            send_single_message(channel, message, message_number=message_counter)
            message_counter+=1
            connection.close()
            input("Press Enter to continue...")

        elif choice == '2':
            consumer_channel, consumer_connection = prepare_consumer_connection()
            consumer_thread = threading.Thread(target=start_consumer, args=(consumer_channel, consumer_connection))
            consumer_thread.start()
            print("Consumer started in background ...\n")

        elif choice == '3':
            if consumer_thread and consumer_thread.is_alive():
                consumer_connection.close()
                consumer_thread.join()
            print("Exiting py-chatter. Goodbye!")
            sys.exit(0)
            break

if __name__ == "__main__":

    # CLI entry point

    print("Welcome to py-chatter")
    try:
        main_menu()
    except KeyboardInterrupt:
        print("\nProgram terminated by user.")
        sys.exit(0)
    except Exception as e:
        print(f"\nAn unexpected error occurred: {e}")
        sys.exit(1)


    # # start producer in separate thread
    # channel, connection = prepare_producer_connection()
    # producer_thread = threading.Thread(target=send_messages_every, args=(connection, channel, 2, 10))
    # producer_thread.start()

    # # start consumer in other thread
    # channel_c, connection_c = prepare_consumer_connection()
    
    # consumer_thread = threading.Thread(target=start_consumer, args=(channel_c, connection_c))
    # consumer_thread.start()

    # producer_thread.join()      
    # consumer_thread.join()

