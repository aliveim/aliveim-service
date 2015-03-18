# aliveim-service

Micro service that spawns Timer based objects and trigger a notification at timeout

# About alive.im

This service aims to solve a problem people may have with many connected devices or servers. Each device "pings" alive.im service just to tell it "Hey! I'm alive!". The service internally keep track of all these devices and their configuration, triggering a notification if the given device doesn't send an "I'm alive!" message in a given time frame. For example you can setup a RaspberryPi device to send an "I'm alive!" message every 1 minute and tell alive.im service to send you an email if it doesn't receive a message within 5 minutes. When the message is received, the timer is reset.

# About alive-im micro service

The micro service listens to an internal interface for a POST message sent by the main service. The main service will communicate the device ID (that will be used to identify the device in the timers pool) and the maximum timeout. Once received these informations, the micro service spawns a Timer based object with the device informations. If the device send another "ping" within the maximum timeout, the Timer is reset. If the timeout expires before receiving a new "ping" message, the notification is triggered.

### Why Go?

Technically I could have implemented everything in Python (using Celery and RabbitMQ for async tasks). I chose Go mainly because I wanted a concrete (but not too complex) project to improve my Go language skills. The other (but not less important) reason is because if the service is going to have thousand of devices "connected" and spawn a Timer based object for each device, it's nice to have a more optimized application that will consume less memory.

# Workflow of the code

* A POST is received by aliveim-service with the following structure
```
{
  'device_id': 'abcd1234',
  'timeout': 300
}
```
* aliveim-service searches inside an **hash table** where the key is **device_id** and the value is a pointer to a Timer based object.
  * If the key is **not found**, a new Timer based object is created with a timeout given by the initial call. An item is added to the hash table with device_id as key and just created Timer based object pointer as value.
  * If the key is **found**, the related Timer based object is get from the hash table and its timeout is reset.
* If one of the Timer based objects reaches the timeout, aliveim-service triggers a notification on the backend, sending the device_id that is involved in the timeout.
