from gpiozero import LED
from time import sleep
import signal

led = LED(3)

while True:
    led.on()
    sleep(0.25)
    led.off()
    sleep(0.25)
    print("blink")

def handle_exit(sig, frame):
    led.off()
    print("bye")

signal.signal(signal.SIGTERM, handle_exit)
signal.signal(signal.SIGINT, handle_exit)
