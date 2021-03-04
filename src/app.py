import logging
import os
from flask import Flask, jsonify, request, Response
from models.elasticsearch import Elasticsearch
from dotenv import load_dotenv
import yaml
import secrets
from models.cache import Cache
from Crypto import Random
from Crypto.Cipher import AES
import base64
from hashlib import md5

load_dotenv()

PASSWORD_LENGTH = 13
BLOCK_SIZE = 16


def trans(key):
    return md5(key).digest()


def encrypt(message, passphrase):
    passphrase = trans(passphrase)
    IV = Random.new().read(BLOCK_SIZE)
    aes = AES.new(passphrase, AES.MODE_CFB, IV)
    return base64.b64encode(IV + aes.encrypt(message))


def decrypt(encrypted, passphrase):
    passphrase = trans(passphrase)
    encrypted = base64.b64decode(encrypted)
    IV = encrypted[:BLOCK_SIZE]
    aes = AES.new(passphrase, AES.MODE_CFB, IV)
    return aes.decrypt(encrypted[BLOCK_SIZE:])


app = Flask(__name__)
app.logger.setLevel(logging.DEBUG)
app.config['SECRET_KEY'] = os.environ.get("SECRET_KEY", default="")

with open(os.getenv("CONFIG_PATH", ''), 'r') as stream:
    try:
        app.config['config'] = yaml.safe_load(stream)
    except yaml.YAMLError as e:
        app.logger.exception(e)


@app.route('/health')
def health():
    return jsonify({
        "status": "OK"
    }), 200


@app.route('/')
def check_user():

    if not request.headers.get("Remote-User"):
        return {
            "name": "Kibana Auth Proxy",
            "info": "Please provide required headers",
        }

    cache = Cache(
        os.getenv("REDIS_HOST", 'localhost'),
        os.getenv("REDIS_PORT", 6379),
        os.getenv("REDIS_DB", 0),
        os.getenv("REDIS_EXPIRE_SECONDS", 3600),
    )

    cache_key = "test-kibana-proxy-auth-{}".format(request.headers.get("Remote-User"))

    print(cache.exists(cache_key))

    if not cache.exists(cache_key):
        password = os.getenv("KIBANA_USER_PASSWORD", secrets.token_urlsafe(PASSWORD_LENGTH))

        auth = (os.getenv("ELASTICSEARCH_USER", ''), os.getenv("ELASTICSEARCH_PASSWORD", ''))

        verify_ssl = False if os.getenv("VERIFY_SSL", '1') == "0" else True

        try:
            elastic = Elasticsearch(os.getenv("ELASTICSEARCH_HOST", ''), verify_ssl, app.logger, auth)
        except Exception as e:
            return {
                "error": str(e),
            }, 500

        # elastic.check_user(request.headers.get("Remote-User"))

        user_groups = request.headers.get("Remote-Groups").split(",")

        roles = [
            app.config['config']['default_role']
        ]
        if app.config['config']['group_mappings']:
            roles = []

            for group in user_groups:
                if group in app.config['config']['group_mappings']:
                    for mapping in app.config['config']['group_mappings'][group]:
                        roles.append(mapping)

        elastic.update_user(
            request.headers.get("Remote-User"),
            password,
            request.headers.get("Remote-Email"),
            request.headers.get("Remote-Name"),
            {
                "groups": user_groups
            },
            roles,
        )

        encrypted_password = encrypt(bytes(password, encoding='utf-8'), bytes(app.config['SECRET_KEY'], encoding='utf-8'))
        cache.set(cache_key, encrypted_password)
        app.logger.debug("Password generated for {}".format(request.headers.get("Remote-User")))

    resp = Response()

    decrypted_pass = decrypt(cache.get(cache_key), bytes(app.config['SECRET_KEY'], encoding='utf-8')).decode("utf-8")
    user_and_pass_string = "{}:{}".format(request.headers.get("Remote-User"), decrypted_pass)
    user_and_pass = base64.b64encode(bytes(user_and_pass_string, encoding='utf-8')).decode("ascii")

    resp.headers = dict(request.headers)
    resp.headers['Authorization'] = "Basic {}".format(user_and_pass)

    return resp
