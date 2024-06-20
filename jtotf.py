import tensorflow as tf
import json
import numpy as np

# Load the JSON data
json_data = {
  "layers": {
    "hidden": [
      {
        "neurons": {
          "n8": {
            "activationType": "tanh",
            "bias": -0.17470446088956545,
            "connections": { "n10": { "weight": -0.14941715679306145 } }
          }
        }
      },
      {
        "neurons": {
          "4": {
            "activationType": "relu",
            "bias": -1.352357620161588,
            "connections": {
              "1": { "weight": 0.671525852096607 },
              "4": { "weight": -0.24317318204605914 },
              "n12": { "weight": -1.5490626847288573 },
              "n9": { "weight": 0.9518683604529907 }
            }
          },
          "n10": {
            "activationType": "swish",
            "bias": 0.3983233997581777,
            "connections": { "n9": { "weight": 0.1857125570934386 } }
          }
        }
      },
      {
        "neurons": {
          "n11": {
            "activationType": "swish",
            "bias": -1.2141084847620203,
            "connections": {}
          }
        }
      },
      {
        "neurons": {
          "n9": {
            "activationType": "elu",
            "bias": -0.38884268171498015,
            "connections": {}
          }
        }
      }
    ],
    "input": {
      "neurons": {
        "1": { "activationType": "", "bias": 0 },
        "2": { "activationType": "", "bias": 0 },
        "3": { "activationType": "", "bias": 0 }
      }
    },
    "output": {
      "neurons": {
        "5": {
          "activationType": "swish",
          "bias": 0.9132838377117213,
          "connections": { "4": { "weight": 1.2938737566724972 } }
        },
        "6": {
          "activationType": "sigmoid",
          "bias": 0.7393680903347826,
          "connections": { "4": { "weight": 1.1112143870402664 } }
        },
        "7": {
          "activationType": "sigmoid",
          "bias": 0.9131860943447563,
          "connections": { "4": { "weight": -0.20286658225128207 } }
        }
      }
    }
  }
}


# Parse the JSON data
layers = json_data['layers']

# Create input layer
inputs = tf.keras.Input(shape=(3,), name='input_layer')  # Assuming 3 input neurons

# Dictionary to store all neurons
neurons = {}

# Add input neurons to the dictionary
for neuron_id in layers['input']['neurons']:
    neurons[neuron_id] = inputs

# Function to get or create a neuron
def get_or_create_neuron(neuron_id, neuron_data):
    if neuron_id not in neurons:
        activation = neuron_data.get('activationType', None)
        if activation == '':
            activation = None
        x = tf.keras.layers.Dense(1, 
                                  activation=activation,
                                  use_bias=True,
                                  name=f'dense_{neuron_id}')(inputs)
        neurons[neuron_id] = x
    return neurons[neuron_id]

# Create all neurons and their connections
for layer in layers['hidden'] + [layers['output']]:
    for neuron_id, neuron_data in layer['neurons'].items():
        neuron = get_or_create_neuron(neuron_id, neuron_data)
        if neuron_data['connections']:
            neuron_inputs = []
            for conn in neuron_data['connections']:
                conn_data = layers['hidden'][0]['neurons'].get(conn, {})
                if not conn_data:
                    for hidden_layer in layers['hidden']:
                        conn_data = hidden_layer['neurons'].get(conn, {})
                        if conn_data:
                            break
                if not conn_data:
                    conn_data = layers['input']['neurons'].get(conn, {})
                neuron_inputs.append(get_or_create_neuron(conn, conn_data))
            x = tf.keras.layers.Concatenate()(neuron_inputs) if len(neuron_inputs) > 1 else neuron_inputs[0]
            neurons[neuron_id] = tf.keras.layers.Dense(1, 
                                                       activation=neuron_data['activationType'],
                                                       use_bias=True,
                                                       name=f'dense_{neuron_id}_connected')(x)

# Get output neurons
outputs = [neurons[neuron_id] for neuron_id in layers['output']['neurons']]

# Combine output neurons
if len(outputs) > 1:
    output = tf.keras.layers.Concatenate()(outputs)
else:
    output = outputs[0]

# Create the model
model = tf.keras.Model(inputs=inputs, outputs=output)

# Set the weights
for layer_data in layers['hidden'] + [layers['output']]:
    for neuron_id, neuron_data in layer_data['neurons'].items():
        layer_name = f'dense_{neuron_id}_connected' if neuron_data['connections'] else f'dense_{neuron_id}'
        try:
            layer = model.get_layer(layer_name)
            weights = []
            bias = np.array([neuron_data['bias']])
            
            if neuron_data['connections']:
                for conn in neuron_data['connections']:
                    weight = neuron_data['connections'][conn]['weight']
                    weights.append([weight])
            else:
                weights = [[0.0]]  # Default weight if no connections
            
            weights = np.array(weights)
            layer.set_weights([weights, bias])
        except ValueError:
            print(f"Warning: Layer {layer_name} not found in the model")

# Print model summary
model.summary()


test_input = np.array([
    [1, 2, 3],
    [2, 3, 4],
    [3, 4, 5]
])

# Use the model to make a prediction
prediction = model.predict(test_input)

# Print the prediction
print("Input:", test_input)
print("Output:", prediction)