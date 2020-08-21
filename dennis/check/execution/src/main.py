#!/usr/bin/python3
import os

from formula import formula

context = os.environ.get("CONTEXT")
executionId = os.environ.get("EXECUTION_ID")
username = os.environ.get("USERNAME")
password = os.environ.get("PASSWORD")
formula.Run(context, executionId, username, password)
