# build.yaml

```yaml
--- # SensorBee plug-in list
plugins:
- pfi/sensorbee/snowflake/plugin
- pfi/sensorbee/pystate/mlstate/plugin
- pfi/sensorbee/pystate/mlstate/example/mnist/plugin
```

# Sample BQL

Download train and test file from [MNIST](http://yann.lecun.com/exdb/mnist/) and unzip.

## Train Phase

Weight matrix is cached in "ml_mnist" state.

```sql
-- create training data source
CREATE PAUSED SOURCE mnist_data TYPE mnist_source
    WITH images_file_name='train-images-idx3-ubyte',
         labels_file_name='train-labels-idx1-ubyte',
         data_size=60000,
         image_element_size=784, -- optional
         rondom=true -- optional
;

-- create multiple cassification state
CREATE STATE ml_mnist TYPE pymlstate
    WITH module_path='',
         module_name='mnist', -- mnist.py
         class_name='MNIST',
         batch_train_size=100
;
CREATE SINK ml_mnist_trainer TYPE uds
    WITH name='ml_mnist'
;
-- training (using shared sink)
INSERT INTO ml_mnist_trainer SELECT RSTREAM
    *
    FROM mnist_data [RANGE 1 TUPLES]
;

RESUME SOURCE mnist_data;

-- total 5 epoch
REWIND SOURCE mnist_data;
REWIND SOURCE mnist_data;
REWIND SOURCE mnist_data;
REWIND SOURCE mnist_data;
```

## Evaluate Phase

Generate test data stream.

```sql
-- create test data sxource
CREATE PAUSED SOURCE mnist_test_data TYPE mnist_source
    WITH images_file_name='t10k-images-idx3-ubyte',
         labels_file_name='t10k-labels-idx1-ubyte',
         data_size=10000,
         image_element_size=784, -- optional
         rondom=false
;

CREATE STREAM ml_mnist_eval AS SELECT RSTREAM
    pymlstate_predict('ml_mnist', me:data) AS pred,
    me:label AS label
    FROM mnist_test_data [RANGE 1 TUPLES] AS me
;
```

### Use saved FunctionSet

`pymlstate` can use trained model, create another state which read saved model and use the state in evaluation stream.

```sql
-- load traned data
CREATE STATE ml_mnist_trained TYPE pymlstate
    WITH module_path='',
         module_name='mnist', -- mnist.py
         class_name='MNIST',
         model_file_path='mnist_model.pkl' -- saved model file path
;

CREATE STREAM ml_mnist_eval AS SELECT RSTREAM
    pymlstate_predict('ml_mnist_trained', me:data) AS pred,
    me:label AS label
    FROM mnist_test_data [RANGE 1 TUPLES] AS me
;
```

Evaluate

```sql
-- logging accuracy
SELECT RSTREAM count(me:pred) / 10000.0
    FROM ml_mnist_eval [RANGE 10000 TUPLES] AS me
    WHERE me:pred = me:label
;

-- call RESUME in other shell
RESUME SOURCE mnist_test_data;
```

* [TODO] these query output 10000 lines, should be modified.
