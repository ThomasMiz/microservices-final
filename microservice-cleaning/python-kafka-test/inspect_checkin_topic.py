from kafka import KafkaConsumer
import json

topic = "check-in"

# Connect to Kafka broker running on localhost:9092
consumer = KafkaConsumer(
    topic,
    bootstrap_servers=['localhost:9092'],
    value_deserializer=lambda m: json.loads(m.decode('utf-8')),
    auto_offset_reset='earliest',
    group_id='inspect-group'
)

print(f"Listening for messages on '{topic}'... (Press Ctrl+C to stop)")

try:
    for message in consumer:
        print(f"\n[{message.timestamp}] Received message:")
        print(f"  Topic: {message.topic}")
        print(f"  Partition: {message.partition}")
        print(f"  Offset: {message.offset}")
        print(f"  Value: {message.value}")
except KeyboardInterrupt:
    print("\n\nStopping consumer...")
finally:
    consumer.close()
