/*
 * SPDX-License-Identifier: Apache-2.0
 */

import {Context, Contract, Info, Returns, Transaction} from "fabric-contract-api";
import {PostOffice} from "./post";

@Info({title: "AssetDelivery", description: "Smart contract for delivery asset in post offices"})
export class AssetDelivery extends Contract {
    async beforeTransaction(ctx: Context) {
        console.log(`<<<<Call Chaincode Methode>>>>`);
        const {fcn, params} = ctx.stub.getFunctionAndParameters();
        console.log(`Methode Name: ${fcn}`);
        console.log(`Parameters: ${params}`);
        console.log(`TxID: ${ctx.stub.getTxID()}`);
    }

    @Transaction()
    public async InitLedger(ctx: Context): Promise<void> {
        const names = ["office1", "office2"];
        for (const name of names) {
            const office = await ctx.stub.getState(name);
            if (!office || office.length === 0) {
                await ctx.stub.putState(name, Buffer.from(JSON.stringify({name: name, addresses: []})));
            }
        }
    }

    // addAddress add new address to postOffice
    @Transaction()
    public async addAddress(ctx: Context, name: string, postalId: string) {
        const office = await this.readPostOffice(ctx, name);
        office.addresses.push(postalId);
        await ctx.stub.putState(name, Buffer.from(JSON.stringify(office)));
    }

    // addAddresses add new addresses to postOffice
    @Transaction()
    public async addAddresses(ctx: Context, name: string, postalIdsStr: string) {
        const postalIds = JSON.parse(postalIdsStr);
        const office = await this.readPostOffice(ctx, name);
        office.addresses = office.addresses.concat(postalIds);
        await ctx.stub.putState(name, Buffer.from(JSON.stringify(office)));
    }

    // deliverAsset deliver asset to office
    @Transaction()
    @Returns("string")
    public async deliverAsset(ctx: Context, postalId: string, assetId: string): Promise<string> {
        return await this.deliverAssetOffice1(ctx, postalId, assetId);
    }

    // deliverAssetOffice1 deliver asset in office1 if address not found pass to office2
    @Transaction()
    @Returns("string")
    public async deliverAssetOffice1(ctx: Context, postalId: string, assetId: string): Promise<string> {
        const name = "office1";
        const office = await this.readPostOffice(ctx, name);
        const addressIndex = office.addresses.findIndex((address) => address === postalId);
        if (addressIndex >= 0) {
            await this.addToPostBox(ctx, postalId, assetId);
            return name;
        } else {
            return await this.deliverAssetOffice2(ctx, postalId, assetId);
        }
    }

    // deliverAssetOffice2 deliver asset in office2 if address not found pass to office1
    @Transaction()
    @Returns("string")
    public async deliverAssetOffice2(ctx: Context, postalId: string, assetId: string): Promise<string> {
        const name = "office2";
        const office = await this.readPostOffice(ctx, name);
        const addressIndex = office.addresses.findIndex((address) => address === postalId);
        if (addressIndex >= 0) {
            await this.addToPostBox(ctx, postalId, assetId);
            return name;
        } else {
            return await this.deliverAssetOffice1(ctx, postalId, assetId);
        }
    }

    // ReadPostOffice returns the office stored in the world state with given name.
    @Transaction(false)
    public async readPostOffice(ctx: Context, name: string): Promise<PostOffice> {
        const office = await ctx.stub.getState(name); // get the office from chaincode state
        if (!office || office.length === 0) {
            throw new Error(`The office ${name} does not exist`);
        }
        return JSON.parse(office.toString());
    }

    // addToPostBox add an asset to postbox
    @Transaction(false)
    public async addToPostBox(ctx: Context, postalId: string, assetId: string): Promise<void> {
        const postboxId = `postbox_${postalId}`;
        const postbox = await ctx.stub.getState(postboxId);
        if (!postbox || postbox.length === 0) {
            await ctx.stub.putState(postboxId, Buffer.from(JSON.stringify([assetId])));
        } else {
            const postBoxObj = JSON.parse(postbox.toString());
            postBoxObj.push(assetId);
            await ctx.stub.putState(postboxId, Buffer.from(JSON.stringify(postBoxObj)));
        }
    }

    // readPostBox returns the postbox stored in the world state with given name.
    @Transaction(false)
    public async readPostBox(ctx: Context, postalId: string): Promise<string[]> {
        const postboxId = `postbox_${postalId}`;
        const postbox = await ctx.stub.getState(postboxId);
        if (!postbox || postbox.length === 0) {
            throw new Error(`The postboc ${postalId} does not exist`);
        }
        return JSON.parse(postbox.toString());
    }
}
