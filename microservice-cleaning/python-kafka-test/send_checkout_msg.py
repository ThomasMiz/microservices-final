from kafka import KafkaProducer
import datetime
import json


# Connect to Kafka broker running on localhost:9092
producer = KafkaProducer(
    bootstrap_servers=['localhost:9092'],
    value_serializer=lambda v: json.dumps(v).encode('utf-8')
)

# Send a message to the topic
topic = 'check-out'
message = {
    'roomNumber': '101',
    'guestId': 'Marcelo',
    'timestamp': datetime.datetime.now().astimezone().isoformat()
}

try:
    future = producer.send(topic, value=message)
    record_metadata = future.get(timeout=10)
    print(f"Message sent successfully to {record_metadata.topic} partition {record_metadata.partition} at offset {record_metadata.offset}")
except Exception as e:
    print(f"Error sending message: {e}")
finally:
    producer.close()
