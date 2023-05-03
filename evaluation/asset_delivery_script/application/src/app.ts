/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
import {Contract, Gateway, GatewayOptions} from "fabric-network";
import {writeFileSync} from "fs";
import * as path from "path";
import {buildCCPOrg1, buildWallet} from "./utils/AppUtil";
import {buildCAClient, enrollAdmin, registerAndEnrollUser} from "./utils/CAUtil";
import {addAddresses, deliverAsset, readPostOffice} from "./transactions";
import {getrandomInt, getrandomPostalId} from "./utils/utils";
import {TransactionResult} from "./types";

const channelName = "mychannel";
const chaincodeName = "post";
const mspOrg1 = "Org1MSP";
const walletPath = path.join(__dirname, "wallet");
const org1UserId = "appUser";

// pre-requisites:
// - fabric-sample two organization test-network setup with two peers, ordering service,
//   and 2 certificate authorities
//         ===> from directory /fabric-samples/test-network
//         ./network.sh up createChannel -ca
// - Use any of the asset-transfer-basic chaincodes deployed on the channel "mychannel"
//   with the chaincode name of "basic". The following deploy command will package,
//   install, approve, and commit the javascript chaincode, all the actions it takes
//   to deploy a chaincode to a channel.
//         ===> from directory /fabric-samples/test-network
//         ./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-typescript/ -ccl javascript
// - Be sure that node.js is installed
//         ===> from directory /fabric-samples/asset-transfer-basic/application-typescript
//         node -v
// - npm installed code dependencies
//         ===> from directory /fabric-samples/asset-transfer-basic/application-typescript
//         npm install
// - to run this test application
//         ===> from directory /fabric-samples/asset-transfer-basic/application-typescript
//         npm start

// NOTE: If you see  kind an error like these:
/*
    2020-08-07T20:23:17.590Z - error: [DiscoveryService]: send[mychannel] - Channel:mychannel received discovery error:access denied
    ******** FAILED to run the application: Error: DiscoveryService: mychannel error: access denied

   OR

   Failed to register user : Error: fabric-ca request register failed with errors [[ { code: 20, message: 'Authentication failure' } ]]
   ******** FAILED to run the application: Error: Identity not found in wallet: appUser
*/
// Delete the /fabric-samples/asset-transfer-basic/application-typescript/wallet directory
// and retry this application.
//
// The certificate authority must have been restarted and the saved certificates for the
// admin and application user are not valid. Deleting the wallet store will force these to be reset
// with the new certificate authority.
//

/**
 *  A test application to show basic queries operations with any of the asset-transfer-basic chaincodes
 *   -- How to submit a transaction
 *   -- How to query and check the results
 *
 * To see the SDK workings, try setting the logging to show on the console before running
 *        export HFC_LOGGING='{"debug":"console"}'
 */
async function getContact(): Promise<Contract> {
    try {
        // build an in memory object with the network configuration (also known as a connection profile)
        const ccp = buildCCPOrg1();

        // build an instance of the fabric ca services client based on
        // the information in the network configuration
        const caClient = buildCAClient(ccp, "ca.org1.example.com");

        // setup the wallet to hold the credentials of the application user
        const wallet = await buildWallet(walletPath);

        // in a real application this would be done on an administrative flow, and only once
        await enrollAdmin(caClient, wallet, mspOrg1);

        // in a real application this would be done only when a new user was required to be added
        // and would be part of an administrative flow
        await registerAndEnrollUser(caClient, wallet, mspOrg1, org1UserId, "org1.department1");

        // Create a new gateway instance for interacting with the fabric network.
        // In a real application this would be done as the backend server session is setup for
        // a user that has been verified.
        const gateway = new Gateway();

        const gatewayOpts: GatewayOptions = {
            wallet,
            identity: org1UserId,
            discovery: {enabled: true, asLocalhost: true}, // using asLocalhost as this gateway is using a fabric network deployed locally
        };

        // setup the gateway instance
        // The user will now be able to create connections to the fabric network and be able to
        // submit transactions and query. All transactions submitted by this gateway will be
        // signed by this user using the credentials stored in the wallet.
        await gateway.connect(ccp, gatewayOpts);

        // Build a network instance based on the channel where the smart contract is deployed
        const network = await gateway.getNetwork(channelName);

        // Get the contract from the network.
        const contract = network.getContract(chaincodeName);
        // Initialize a set of asset data on the channel using the chaincode 'InitLedger' function.
        // This type of transaction would only be run once by an application the first time it was started after it
        // deployed the first time. Any updates to the chaincode deployed later would likely not need to run
        // an "init" type function.
        console.log("\n--> Submit Transaction: InitLedger, function creates the initial set of assets on the ledger");
        await contract.submitTransaction("InitLedger");
        console.log("*** Result: committed");
        return contract;
    } catch (error) {
        console.error(`******** FAILED to run the application: ${error}`);
    }
}

function delay(ms: number, result?: number) {
    return new Promise((resolve) => setTimeout(() => resolve(result), ms));
}

async function addAddressToPostOffice(contract: Contract, count: number, officeIndex: number): Promise<void> {
    const office = await readPostOffice(contract, `office${officeIndex}`);
    const postalIds: string[] = office.addresses;
    const newIds: string[] = [];
    for (let index = 0; index < count; index++) {
        const newId = getrandomPostalId(postalIds, 999999, `${officeIndex}`);
        postalIds.push(newId);
        newIds.push(newId);
    }
    await addAddresses(contract, `office${officeIndex}`, newIds);
    console.log(`--> Submit Transaction: addAddress, count: ${count}`);
}

const deliverAssetsValidAddress = (contract: Contract, count: number, addresses: string[], delayTime: number) => {
    return new Promise((resolve: (results: TransactionResult[]) => void) => {
        const promisses: Promise<TransactionResult>[] = [];
        let index = 0;
        const interval = setInterval(() => {
            if (index >= count) {
                Promise.all(promisses).then((results) => {
                    resolve(results);
                });
                clearInterval(interval);
            } else {
                if (index % 20 === 15) {
                    promisses.push(deliverAsset(contract, `3${getrandomInt(10000000)}`, `${getrandomInt(20000)}`));
                } else {
                    promisses.push(deliverAsset(contract, addresses[index], `${getrandomInt(20000)}`));
                }
                index++;
            }
        }, delayTime);
    });
};

async function main() {
    const contract = await getContact();
    // await addAddressToPostOffice(contract, 2000, 1);
    // await addAddressToPostOffice(contract, 2000, 2);
    const office1 = await readPostOffice(contract, "office1");
    const office2 = await readPostOffice(contract, "office2");
    const addresses = office1.addresses.concat(office2.addresses);

    const trasactionCounts: number[] = [100, 200, 400, 600, 800, 1000, 2000];
    for (let index = 0; index < trasactionCounts.length; index++) {
        const count = trasactionCounts[index];
        let results = await deliverAssetsValidAddress(contract, count, addresses, 50);
        let totalTimes = 0;
        for (let index = 0; index < results.length; index++) {
            totalTimes += results[index].time;
        }
        console.log("------------------------------------");
        console.log(`Count = ${count}`);
        console.log(`Average = ${totalTimes / results.length}`);
        console.log(`Sum = ${totalTimes}`);
        const data = {results, avg: totalTimes / results.length, sum: totalTimes};
        writeFileSync(`results/results-invalid-${count}.json`, JSON.stringify(data));
    }
}

main();
