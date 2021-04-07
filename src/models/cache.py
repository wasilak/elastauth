import redis


class Cache(object):

    def __init__(self, host, port, db, time_to_live):
        self.redis = redis.StrictRedis(
            host=host,
            port=port,
            db=db,
        )
        self.time_to_live = int(time_to_live)

    def exists(self, key):
        return self.redis.exists(key) > 0

    def get(self, key):
        return self.redis.get(key).decode("utf-8")

    def set(self, key, value):
        return self.redis.set(key, value, self.time_to_live)

    def ttl(self, key):
        return self.redis.ttl(key)

    def expire(self, key):
        return self.redis.expire(key, self.time_to_live)
