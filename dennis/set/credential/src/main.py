#!/usr/bin/python3
import os

from formula import formula

username = os.environ.get("USERNAME")
password = os.environ.get("PASSWORD")
provider = os.environ.get("PROVIDER")

formula.Run(username, password, provider)
