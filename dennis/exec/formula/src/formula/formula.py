#!/usr/bin/python3

from pick import pick
from termcolor import cprint
from datetime import datetime
import socket
import getpass
import requests
import sys
import json
import uuid
import time

def Run(username, password):

    loginToken = login(username, password)
    formulaResponse = formula(loginToken)    

    options = []
    for ctx in formulaResponse['contexts']:
        options.append(ctx['name'])
    selectedCtx, indexCtx = pick(options, 'Select a context: ', indicator='=>')

    options = []
    for ctx in formulaResponse['formulas']:
        options.append(ctx['command'])
    selectedFormName, indexForm = pick(options, 'Select a formula: ', indicator='=>')

    selectedForm = formulaResponse['formulas'][indexForm]

    inputTypes = {
        "text": textType,
        "password": passwordType,
        "bool": boolType
    }

    requestInputs = []

    for formInputs in selectedForm['inputs']:
        inputType = formInputs['type']
        if(inputType in inputTypes):
            inputTypes[inputType](formInputs, requestInputs)
        else:
            inputCredential(formInputs, requestInputs)

    addParam("IP", "text", getLocalIp(), requestInputs)
    id = str(uuid.uuid4())
    commandData = {"id": id, "command": selectedForm['command'], "inputs": requestInputs}

    respCommand = requests.post(Const.host+'/commands', 
                            json=commandData,
                            headers={
                                'Content-Type':'application/json', 
                                'x-org':'zup', 
                                'x-authorization': loginToken,
                                'x-ctx': selectedCtx})

    serviceResp = {
        201: http200,
        401: http4O1,
        403: http403
    }

    if(respCommand.status_code in serviceResp):
        serviceResp[respCommand.status_code](id, respCommand)
    else:
        httpError(id, respCommand)

    startTime = datetime.now()

    while True:

        currentTime = datetime.now()
        if (currentTime-startTime).total_seconds() >= 60:
            print("Your request is being processed. You can check the execution with the command [rit dennis check execution]")
            print("Execution ID: {}".format(id))
            print("Execution context: {}".format(selectedCtx))
            break

        print("Awaiting execution.", end="", flush=True)
        time.sleep(1)
        print(".", end="", flush=True)
        time.sleep(1)
        print(".", end="", flush=True)
        time.sleep(1)
        print(".", end="", flush=True)
        time.sleep(1)
        print(".", end="", flush=True)
        time.sleep(1)
        print(".")

        if checkExecution(loginToken, selectedCtx, id):
            break


def checkExecution(loginToken, selectedCtx, id):
    print("Checking execution status...")
    
    respCheckExec = requests.get(Const.host+'/executions/'+id, 
                            headers={
                                'Content-Type':'application/json', 
                                'x-org':'zup', 
                                'x-authorization': loginToken,
                                'x-ctx': selectedCtx})

    respExecData = respCheckExec.json()

    if respCheckExec.status_code != 200 or respExecData['status'] != 'Ready':
        print("Execution not found or it's being processed\n")
        return False

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

    return True


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


def formula(loginToken):

    print("Loading context and formulas...")

    respFormula = requests.get(Const.host+'/formulas', 
                            headers={
                                'Content-Type':'application/json', 
                                'x-org':'zup', 
                                'x-authorization': loginToken})

    if respFormula.status_code != 200:
        print('POST /login {}'.format(respFormula.status_code))
        sys.exit()

    return respFormula.json()

def textType(ipt, requestInputs):
    v = input(ipt['label'])
    addParamInput(ipt, requestInputs, v)

def passwordType(ipt, requestInputs):
    v = getpass.getpass(prompt=ipt['label'])
    addParamInput(ipt, requestInputs, v)

def boolType(ipt, requestInputs):
    v = str(bool(input(ipt['label']))).lower()
    addParamInput(ipt, requestInputs, v)

def addParamInput(ipt, requestInputs, v):
    addParam(ipt['name'], ipt['type'], v, requestInputs)

def addParam(name, type, value, requestInputs):
    p = ParamRequest(name, type, value)
    requestInputs.append(p.json())

def addCredential(ipt, requestInputs):
    c = CredentialRequest(ipt['name'], ipt['type'])
    requestInputs.append(c.json())

def getLocalIp():
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    s.connect((Const.IP888, Const.port80))
    localIp = s.getsockname()[0]
    s.close()
    return localIp

def inputCredential(formInputs, requestInputs):
    if(formInputs['type'].startswith('CREDENTIAL_')):
        addCredential(formInputs, requestInputs)


def http200(id, response):
    cprint("Formula running with id [{}]".format(id), 'green')

def http4O1(id, response):
    cprint('user not logged in', 'red')

def http403(id, response):
    cprint("user not unauthorized", 'red')

def httpError(id, response):
    cprint("command failed", 'red')
    sys.exit()

class Const:
    host = 'https://dennis.devdennis.zup.io'
    port80 = 80
    IP888 = "8.8.8.8"


class ParamRequest:
  def __init__(self, name, type, value):
    self.name = name
    self.type = type
    self.value = value

  def json(self):
      return {"name": self.name, "type": self.type, "value": self.value}

class CredentialRequest:
  def __init__(self, name, type):
    self.name = name
    self.type = type

  def json(self):
    return {"name": self.name, "type": self.type}