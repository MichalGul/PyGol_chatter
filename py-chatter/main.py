from producer.simple_producer import send_messages_every, prepare_connection

if __name__ == "__main__":
    print("Welcome to py-chatter")
    channel, connection = prepare_connection()
    send_messages_every(connection, channel, 2, 10)