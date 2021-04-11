import os
import requests
import time
import _thread as thread
import websocket
import RPi.GPIO as GPIO

GPIO.setwarnings(False)
GPIO.setmode(GPIO.BOARD)

FMS_IP = "10.0.100.05"
FMS_PORT = "8080"
FMS_SERVER = FMS_IP + ":" + FMS_PORT
#ALLIANCE_COLOR = 'red' # Change accordingly
ALLIANCE_COLOR = 'blue' # Change accordingly
USERNAME = 'admin'
PASSWORD = 'Password1'

goal_char_msg_map = {
    "I": '{ "type": "CI" }',
    "O": '{ "type": "CO" }',
    "L": '{ "type": "CL" }'
}

#Global Counters
innerCount = 0
outerCount = 0
lowerCount = 0

#GPIO Assinments
innerCount_Pin = 8
outerCount_Pin = 7
lowerCount_Pin = 3

# Inerner Count Input setup
GPIO.setup(innerCount_Pin,GPIO.IN,pull_up_down=GPIO.PUD_DOWN)

def inner_callback(channel):
    global innerCount
    innerCount += 1
    
GPIO.add_event_detect(innerCount_Pin,GPIO.RISING,callback=inner_callback, bouncetime=200)

# Outer Count Input setup
GPIO.setup(outerCount_Pin,GPIO.IN,pull_up_down=GPIO.PUD_DOWN)

def outer_callback(channel):
    global outerCount
    outerCount += 1
    
GPIO.add_event_detect(outerCount_Pin,GPIO.RISING,callback=outer_callback, bouncetime=200)

# Lower Count Input setup
GPIO.setup(lowerCount_Pin,GPIO.IN,pull_up_down=GPIO.PUD_DOWN)

def lower_callback(channel):
    global lowerCount
    lowerCount += 1
    
GPIO.add_event_detect(lowerCount_Pin,GPIO.RISING,callback=lower_callback, bouncetime=200)

#Function to wait for a Power Cell to be scored
def get_Power_Cell_to_Count():
    global innerCount
    global outerCount
    global lowerCount
    while(outerCount == 0 and innerCount == 0 and lowerCount == 0):
        one = 1
    return True

#Retrieve the keyStroke char for the goal to be tallied
def get_msg_from_goal_char(goal_char):
    return goal_char_msg_map[goal_char]

def get_on_ws_open_callback():
    def on_ws_open(ws):
        print("Connected to FMS")

        def run(*args):
            while(True):
                global innerCount
                global outerCount
                global lowerCount
                #Check for any counters that need to be tallied
                sendData = get_Power_Cell_to_Count()
                #What Counter?
                if (outerCount > 0):
                    goal_char = "O"
                    outerCount -= 1
                elif (innerCount > 0):
                    goal_char = "I"
                    innerCount -= 1
                elif (lowerCount > 0):
                    goal_char = "L"
                    lowerCount -= 1
                else:
                    print("nothing to Score")
                    goal_char = ""
                    
                    
                print(f'Info: recieved "{goal_char}"')

                if (goal_char in goal_char_msg_map):
                    print(f'Info: sent {get_msg_from_goal_char(goal_char)}')
                    try:
                        ws.send(get_msg_from_goal_char(goal_char))
                    except:
                        open_websocket()
                else:
                    print('Error: unknown char recieved')

        thread.start_new_thread(run, ())
    
    return on_ws_open
    
def open_websocket():
    def reopen_websocket():
        print("attempt to ReOpen Websocket")
        open_websocket()
    
    print("request.post")
    res = requests.post(f'http://{FMS_SERVER}/login'
        , data={'username': USERNAME, 'password': PASSWORD}
        , allow_redirects=False
    )

    print("Create ws")
    ws = websocket.WebSocketApp(f'ws://{FMS_SERVER}/panels/scoring/{ALLIANCE_COLOR}/websocket'
        , on_open=get_on_ws_open_callback()
        , on_close=reopen_websocket
        , cookie="; ".join(["%s=%s" %(i, j) for i, j in res.cookies.get_dict().items()])
    )

    print("Run Forever")
    ws.run_forever()

def main():
    #Wait for Network connection to FMS
    while(True):
        print(f'Check Network Connection {FMS_IP}')
        response = os.system("ping -c 1 " + FMS_IP)
        if response == 0:
            print(f'{FMS_IP} Found')
        else:
            pingstatus = "Network Error"
            print("Network Error")
        if(response == 0): break
        time.sleep(2)
    
    print("Open Web Socket")
    open_websocket()

main()