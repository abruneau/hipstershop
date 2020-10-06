#!/usr/bin/python

import logging
import sys
from pythonjsonlogger import jsonlogger

class CustomJsonFormatter(jsonlogger.JsonFormatter):
  def add_fields(self, log_record, record, message_dict):
    super(CustomJsonFormatter, self).add_fields(log_record, record, message_dict)
    if not log_record.get('timestamp'):
      log_record['timestamp'] = record.created
    if log_record.get('severity'):
      log_record['severity'] = log_record['severity'].upper()
    else:
      log_record['severity'] = record.levelname

def getJSONLogger(name):
  json_handler = logging.StreamHandler(sys.stdout)
  formatter = CustomJsonFormatter('(timestamp) (severity) (name) (message)')
  json_handler.setFormatter(formatter)
  logger = logging.getLogger(name)
  logger.addHandler(json_handler)
  logger.setLevel(logging.INFO)
  return logger
