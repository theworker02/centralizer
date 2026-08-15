"""Example Python analytics target for Centralizer."""


def calculate(value):
    return float(value) * 2.0


def ping():
    return "ok"


def count_up(n):
    for i in range(int(n)):
        yield i
