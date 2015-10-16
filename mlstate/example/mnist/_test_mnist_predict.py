#!/usr/bin/env python
"""Test script for mnist.py, not use SensorBee
Use saved model (pickle) and only prediction.

"""
import argparse

import numpy as np

import data
import mnist as m


parser = argparse.ArgumentParser(description='Chainer example: MNIST')
parser.add_argument('--gpu', '-g', default=-1, type=int,
                    help='GPU ID (negative value indicates CPU)')
args = parser.parse_args()

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
mf = m.MNIST({'model_file_path': 'mnist_model.pkl', 'gpu': args.gpu})

# evaluation
acc_cnt = 0
for i in range(N_test):
    y = mf.predict(x_test[i])

    if y == y_test[i]:
        acc_cnt += 1

print('test  mean accuracy={}'.format(float(acc_cnt) / N_test))
