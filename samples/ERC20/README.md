
Install
=======
Install the chaincode on the Acme Peer-1
  .    set-env.sh    acme
  set-chain-env.sh       -n erc20  -v 1.0   -p  token/ERC20   
  chain.sh install -p

Instantiate
===========
Instantiate the chaincode

 set-chain-env.sh        -c   '{"Args":["init","ACFT","1000", " a CFab Token!!!","sara"]}'
 chain.sh  instantiate

Query
=====
Query the balance for 'sara'
 set-chain-env.sh         -q   '{"Args":["balanceOf","sara"]}'
 chain.sh query

Invoke
======
Transfer 100 tokens from 'sara' to 'sam'
  set-chain-env.sh         -i   '{"Args":["transfer", "sara", "sam", "10"]}'
  chain.sh  invoke

Query
=====
Check the balance for 'sara' & 'sam'
 set-chain-env.sh         -q   '{"Args":["balanceOf","sara"]}'
 chain.sh query
 set-chain-env.sh         -q   '{"Args":["balanceOf","sam"]}'
 chain.sh query


Events 
==============
Launch the events utility
 events.sh -t chaincode -n erc20 -e transfer -c airlinechannel 

In a <<Terminal #2>  invoke -  transfer events in terminal 1
  set-chain-env.sh         -i   '{"Args":["transfer", "sara", "sam", "10"]}'
  chain.sh invoke

