FROM golang:1.18

RUN apt update
RUN apt install -y pkg-config libopus-dev libopusfile-dev unzip

RUN git clone https://github.com/snowzhop/go-s2s-bot.git

WORKDIR /go/go-s2t-bot/internal/vosk/local

RUN wget https://github.com/alphacep/vosk-api/releases/download/v0.3.42/vosk-linux-x86_64-0.3.42.zip
RUN unzip ./vosk-linux-x86_64-0.3.42.zip

RUN wget https://alphacephei.com/vosk/models/vosk-model-small-ru-0.22.zip
RUN unzip ./vosk-model-small-ru-0.22.zip

ENV VOSK_MODEL_PATH=/go/go-s2t-bot/internal/vosk/local/vosk-model-small-ru-0.22

ENV VOSK_PATH=/go/go-s2t-bot/internal/vosk/local/vosk-linux-x86_64-0.3.42

ARG TGBOTAPI_TOKEN_IN=""
ENV TGBOTAPI_TOKEN=${TGBOTAPI_TOKEN_IN}

WORKDIR /go/go-s2s-bot/cmd/bot_v2

RUN LD_LIBRARY_PATH=${VOSK_PATH} CGO_CPPFLAGS="-I ${VOSK_PATH}" CGO_LDFLAGS="-L ${VOSK_PATH}" \
    go build -v -o /usr/local/bin/app

ENV LD_LIBRARY_PATH=${VOSK_PATH}
ENV CGO_CPPFLAGS="-I ${VOSK_PATH}"
ENV CGO_LDFLAGS="-L ${VOSK_PATH}"

CMD LD_LIBRARY_PATH=${VOSK_PATH} CGO_CPPFLAGS="-I ${VOSK_PATH}" CGO_LDFLAGS="-L ${VOSK_PATH}" go run main.go