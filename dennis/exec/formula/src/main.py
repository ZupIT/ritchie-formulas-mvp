#!/usr/bin/python3
import os

from formula import formula

username = os.environ.get("USERNAME")
password = os.environ.get("PASSWORD")

formula.Run(username, password)
