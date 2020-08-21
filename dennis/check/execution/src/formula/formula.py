#!/usr/bin/python3

import requests
import time
from termcolor import cprint

def Run(context, executionId, username, password):
    
    loginToken = login(username, password)

    print("Checking execution status...")

    respCheckExec = requests.get(Const.host+'/executions/'+executionId, 
                            headers={
                                'Content-Type':'application/json', 
                                'x-org':'zup', 
                                'x-authorization': loginToken,
                                'x-ctx': context})

    if respCheckExec.status_code != 200:
        print('GET /executions/{} {}'.format(id, respCheckExec.status_code))
        sys.exit()


    respExecData = respCheckExec.json()

    startTime = time.strftime('%Y-%m-%d %H:%M:%S', time.localtime(respExecData['content']['startTime']))
    endTime = time.strftime('%Y-%m-%d %H:%M:%S', time.localtime(respExecData['content']['endTime']))
    
    print("")
    cprint('Execution log', 'yellow')

    print("Status: {}".format(respExecData['status']))
    print("Start: {}".format(startTime))
    print("End: {}".format(endTime))
    print("User: {}".format(respExecData['content']['user']))
    print("User IP: {}".format(getIp(respExecData['content']['formulaInputs'])))
    print("Stdout: ")
    print(respExecData['content']['formulaErr'])
    print("Output: ")
    print(respExecData['content']['formulaOutput'])

def getIp(formulaInput):
    for ipt in formulaInput:
        if ipt['name'] == 'IP':
            return ipt['value']

def login(username, password):

    print("Authenticating...")

    loginData = {"username": username, "password": password}

    respLogin = requests.post(Const.host+'/login',
                            json=loginData,
                            headers={'Content-Type':'application/json', 'x-org':'zup'})

    if respLogin.status_code != 200:
        print('POST /login {}'.format(respLogin.status_code))
        sys.exit()
    
    return respLogin.json()['token']



class Const:
    host = 'https://dennis.devdennis.zup.io'