#!/usr/bin/env python
"""Chainer example: train a multi-layer perceptron on MNIST

This is a minimal example to write a feed-forward net. It requires scikit-learn
to load MNIST dataset.

"""
import sys
import numpy as np

import chainer
from chainer import cuda, FunctionSet
import chainer.functions as F
from chainer import optimizers


class MNIST(object):

    def __init__(self):
        self.gpu = -1

        n_units = 1000

        self.model = FunctionSet(
            l1=F.Linear(784, n_units),
            l2=F.Linear(n_units, n_units),
            l3=F.Linear(n_units, 10))

        if self.gpu >= 0:
            cuda.init(self.gpu)
            self.model.to_gpu()

        # Setup optimizer
        self.optimizer = optimizers.Adam()
        self.optimizer.setup(self.model.collect_parameters())

    def forward(self, x_data, train=True):
        x = chainer.Variable(x_data)
        h1 = F.dropout(F.relu(self.model.l1(x)),  train=train)
        h2 = F.dropout(F.relu(self.model.l2(h1)), train=train)
        return self.model.l3(h2)

    def fit(self, xys):
        x = []
        y = []
        for d in xys:
            x.append(d['data'])
            y.append(d['label'])
        x_batch = np.array(x, dtype=np.float32)
        y_batch = np.array(y, dtype=np.int32)

        if self.gpu >= 0:
            x_batch = cuda.to_gpu(x_batch)
            y_batch = cuda.to_gpu(y_batch)

        self.optimizer.zero_grads()
        y = self.forward(x_batch)

        t = chainer.Variable(y_batch)
        loss = F.softmax_cross_entropy(y, t)
        acc = F.accuracy(y, t)

        loss.backward()
        self.optimizer.update()

        nloss = float(cuda.to_cpu(loss.data)) * len(y_batch)
        naccuracy = float(cuda.to_cpu(acc.data)) * len(y_batch)

        retmap = {
            'loss': nloss,
            'accuracy': naccuracy,
        }

        return retmap

    def predict(self, x):
        # non batch
        xx = []
        xx.append(x)
        x_data = np.array(xx, dtype=np.float32)
        if self.gpu >= 0:
            x_data = cuda.to_gpu(x_data)

        y = self.forward(x_data, train=False)
        y = y.data.reshape(y.data.shape[0], y.data.size / y.data.shape[0])
        pred = y.argmax(axis=1)

        return int(pred[0])
