FROM locustio/locust
COPY ./locustfile.py /tmp/locustfile.py
CMD [ "--host", "http://frontend:80", "-f", "/tmp/locustfile.py", "--headless" ]
