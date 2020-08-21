#!/usr/bin/python3

from pick import pick
from termcolor import cprint
import getpass
import requests
import sys

def Run(username, password, provider):

    if provider == "github":
        user = input("Username: ") 
        secret = getpass.getpass(prompt='Token: ')
    else:
        user = input("AccessKeyID: ") 
        secret = getpass.getpass(prompt='SecretAccessKey: ')
    
    loginToken = login(username, password)
    formulaResponse = formula(loginToken)

    title = 'Select a context: '
    options = []

    for ctx in formulaResponse['contexts']:
        options.append(ctx['name'])

    selectedCtx, index = pick(options, title, indicator='=>')

    setCredential(provider, user, secret, loginToken, selectedCtx)


def login(username, password):

    loginData = {"username": username, "password": password}

    respLogin = requests.post(Const.host+'/login', 
                            json=loginData,
                            headers={'Content-Type':'application/json', 'x-org':'zup'})

    if respLogin.status_code != 200:
        print('POST /login {}'.format(respLogin.status_code))
        sys.exit()
    
    return respLogin.json()['token']  


def formula(loginToken):

    respFormula = requests.get(Const.host+'/formulas', 
                            headers={
                                'Content-Type':'application/json', 
                                'x-org':'zup', 
                                'x-authorization': loginToken})

    if respFormula.status_code != 200:
        print('POST /login {}'.format(respFormula.status_code))
        sys.exit()

    return respFormula.json()

def setCredential(provider, user, secret, loginToken, selectedCtx):
    if provider == "github":
        credentialData = {
            'service': provider,
            'credential':{
                'username':user,
                'token':secret
            }
        }
    else:
        credentialData = {
            'service': provider,
            'credential':{
                'accesskeyid':user,
                'secretaccesskey':secret
            }
        }

    respCredential = requests.post(Const.host+'/credentials', 
                            json=credentialData,
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

    serviceResp[respCredential.status_code]()

def http200():
    cprint('ok', 'green')

def http4O1():
    cprint('user not logged in', 'red')

def http403():
    cprint("user not unauthorized", 'red')


class Const:
    host = 'https://dennis.devdennis.zup.io'