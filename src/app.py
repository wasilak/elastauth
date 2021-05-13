import logging
import os
from flask import Flask, jsonify, request, Response, render_template
from models.elasticsearch import Elasticsearch, UserCreationState
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


class AppException(Exception):
    """Custom exception class for handling internal errors."""

    pass


class Factory(object):
    """Factory for creating and storing singletons."""

    ELASTICSEARCH_OBJECT = None
    CACHE_OBJECT = None

    def get_cache():
        """Create/return singleton connection to redis."""
        if Factory.CACHE_OBJECT is None:
            Factory.CACHE_OBJECT = Cache(
                os.getenv("REDIS_HOST", 'localhost'),
                os.getenv("REDIS_PORT", 6379),
                os.getenv("REDIS_DB", 0),
                os.getenv("REDIS_EXPIRE_SECONDS", 3600),
            )
        return Factory.CACHE_OBJECT

    def get_elasticsearch(app_obj):
        """Create/return singleton connection to elasticsearch."""
        if Factory.ELASTICSEARCH_OBJECT is None:

            auth = (os.getenv("ELASTICSEARCH_USER", ''), os.getenv("ELASTICSEARCH_PASSWORD", ''))

            verify_ssl = False if os.getenv("VERIFY_SSL", '1') == "0" else True

            Factory.ELASTICSEARCH_OBJECT = Elasticsearch(os.getenv("ELASTICSEARCH_HOST", ''), verify_ssl, app_obj.logger, auth)
        return Factory.ELASTICSEARCH_OBJECT


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


def get_user_attribute(attribute):
    """Check if Remote user attribute header exists and return value."""
    header_name = 'Remote-' + attribute
    val = request.headers.get(header_name)

    required_headers = ["User"]

    if not val and attribute in required_headers:
        raise AppException('Header not provided: ' + header_name)

    return val


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


@app.route('/config')
def config():

    if 'Content-Type' in request.headers:
        if request.headers['Content-Type'] == "text/yaml":
            return Response(yaml.dump(app.config['config'], default_flow_style=False, explicit_start=True, width=float("inf"), line_break=""), mimetype='text/yaml')
        if request.headers['Content-Type'] == "application/json":
            return app.config['config']

    return render_template('config.html', config=yaml.dump(app.config['config']))


@ app.route('/')
def check_user():
    try:
        user = get_user_attribute('User')
        if not user:
            return {
                "name": "Kibana Auth Proxy",
                "info": "Please provide required headers",
            }

        cache = Factory.get_cache()

        cache_key = "elastauth-{}".format(user)

        if not cache.exists(cache_key) or cache.ttl(cache_key) <= 0:
            password = os.getenv("KIBANA_USER_PASSWORD", secrets.token_urlsafe(PASSWORD_LENGTH))

            try:
                elastic = Factory.get_elasticsearch(app)
            except Exception as es_exc:
                raise AppException('Error whilst connecting to elasticsearch: {}'.format(str(es_exc)))

            # elastic.check_user(request.headers.get("Remote-User"))

            user_groups = request.headers.get("Remote-Groups").split(",")

            roles = []

            if app.config['config']['group_mappings']:
                for group in user_groups:
                    if group in app.config['config']['group_mappings']:
                        for mapping in app.config['config']['group_mappings'][group]:
                            roles.append(mapping)

            if len(roles) == 0:
                roles = [
                    app.config['config']['default_role']
                ]

            user_creation_state = elastic.update_user(
                user,
                password,
                get_user_attribute('Email'),
                get_user_attribute('Name'),
                {
                    "groups": user_groups
                },
                roles,
            )
            if user_creation_state is UserCreationState.ERROR:
                raise AppException('Erorr whilst creating/updating user')

            encrypted_password = encrypt(bytes(password, encoding='utf-8'), bytes(app.config['SECRET_KEY'], encoding='utf-8'))
            cache.set(cache_key, encrypted_password)
            app.logger.debug("Password generated for {}".format(user))

        # if user cache is valid, and ELASTAUTH_CACHE_EXTEND == "true" then extend cache TTL without changing password
        if os.getenv("ELASTAUTH_CACHE_EXTEND", "false") == "true" and cache.ttl(cache_key) > 0 and cache.ttl(cache_key) < cache.time_to_live:
            app.logger.debug("Extending cache TTL (from {} to {}): password generated for {}".format(cache.ttl(cache_key), cache.time_to_live, user))
            cache.expire(cache_key)

        resp = Response()

        decrypted_pass = decrypt(cache.get(cache_key), bytes(app.config['SECRET_KEY'], encoding='utf-8')).decode("utf-8")
        user_and_pass_string = "{}:{}".format(user, decrypted_pass)
        user_and_pass = base64.b64encode(bytes(user_and_pass_string, encoding='utf-8')).decode("ascii")

        resp.headers = dict(request.headers)
        resp.headers['Authorization'] = "Basic {}".format(user_and_pass)

        return resp

    except AppException as app_exc:
        return {"error": str(app_exc)}, 500
