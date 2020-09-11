FROM python:3-slim

RUN apt-get -qq update \
    && apt-get install -y --no-install-recommends \
        g++ curl

COPY requirements.txt .

RUN mkdir /install
RUN pip install -r requirements.txt

COPY . .
RUN chmod +x ./loadgen.sh
ENTRYPOINT ./loadgen.sh
