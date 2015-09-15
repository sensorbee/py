#!/usr/bin/env python
"""Test script for mnist.py, not use SensorBee
Fit and predict MNIST data, and save modle as pickle.

"""
import argparse

import numpy as np
import six

import data
import mnist as m


parser = argparse.ArgumentParser(description='Chainer example: MNIST')
parser.add_argument('--gpu', '-g', default=-1, type=int,
                    help='GPU ID (negative value indicates CPU)')
args = parser.parse_args()

batchsize = 100
n_epoch = 10  # chainer example is 20

# Prepare dataset
print('load MNIST dataset')
mnist = data.load_mnist_data()
mnist['data'] = mnist['data'].astype(np.float32)
mnist['data'] /= 255
mnist['target'] = mnist['target'].astype(np.int32)

N = 60000
x_train, x_test = np.split(mnist['data'],   [N])
y_train, y_test = np.split(mnist['target'], [N])
N_test = y_test.size

# Neural net architecture
mf = m.MNIST()

# Learning loop
for epoch in six.moves.range(1, n_epoch + 1):
    print('epoch', epoch)

    # training
    perm = np.random.permutation(N)
    sum_accuracy = 0
    sum_loss = 0
    for i in six.moves.range(0, N, batchsize):
        x_batch = x_train[perm[i:i + batchsize]]
        y_batch = y_train[perm[i:i + batchsize]]
        xys = []
        for i in range(batchsize):
            xy = {
                'data': x_batch[i],
                'label': y_batch[i],
            }
            xys.append(xy)

        ret = mf.fit(xys)

        sum_loss += ret['loss']
        sum_accuracy += ret['accuracy']

    print('train mean loss={}, accuracy={}'.format(
        sum_loss / N, sum_accuracy / N))

    # evaluation
    acc_cnt = 0
    for i in range(N_test):
        y = mf.predict(x_test[i])

        if y == y_test[i]:
            acc_cnt += 1

    print('test  mean accuracy={}'.format(float(acc_cnt) / N_test))

# save model
with open('mnist_model.pkl', 'wb') as output:
    six.moves.cPickle.dump(mf.get_model(), output, -1)
    print('save done')
