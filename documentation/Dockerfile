FROM python:3.13-slim

WORKDIR /docs

RUN pip install --no-cache-dir \
    mkdocs-material \
    pymdown-extensions

COPY . .

EXPOSE 8000

ENTRYPOINT ["mkdocs"]
CMD ["serve", "--dev-addr=0.0.0.0:8000"]
