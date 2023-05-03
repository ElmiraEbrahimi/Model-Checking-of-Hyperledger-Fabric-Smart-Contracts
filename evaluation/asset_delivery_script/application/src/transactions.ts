import {Contract} from "fabric-network";
import {PostOffice, TransactionResult} from "./types";

const addAddresses = async (contract: Contract, name: string, postalIds: string[]): Promise<void> => {
    await contract.submitTransaction("addAddresses", name, JSON.stringify(postalIds));
};

const deliverAsset = async (contract: Contract, postalId: string, assetId: string): Promise<TransactionResult> => {
    const startTime = Date.now();
    try {
        await contract.submitTransaction("deliverAsset", postalId, assetId);
        return {result: "SUCCESS", time: Date.now() - startTime};
    } catch (error) {
        return {result: "FAILED", time: Date.now() - startTime};
    }
};

const readPostOffice = async (contract: Contract, name: string): Promise<PostOffice> => {
    const result = await contract.evaluateTransaction("readPostOffice", name);
    return JSON.parse(result.toString());
};

const readPostBox = async (contract: Contract, postalId: string): Promise<PostOffice> => {
    const result = await contract.evaluateTransaction("readPostBox", postalId);
    return JSON.parse(result.toString());
};

export {deliverAsset, readPostOffice, readPostBox, addAddresses};
