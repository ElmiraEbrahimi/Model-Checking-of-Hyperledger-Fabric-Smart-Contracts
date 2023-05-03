const colors = ["blue", "red", "green", "yellow", "black", "brown", "white", "pink", "orange"];
const owners = ["Saeed", "Ali", "vahid", "Hojjat", "Bahar", "Farzaneh", "Mina", "Sam", "Goli", "Reza", "Hamid", "Hamed", "Fariba", "Romina", "Hana"];
const domains = ["yahoo.com", "gmail.com", "hotmail.com", "systemgroup.net", "aol.com", "hotmail.co.uk", "hotmail.fr", "msn.com", "ut.ir"];

const getrandomInt = (max: number, min: number = 0): number => {
    min = Math.ceil(min);
    max = Math.floor(max);
    return Math.floor(Math.random() * (max - min + 1)) + min;
};

const getrandomColor = (): string => {
    return colors[getrandomInt(colors.length - 1)];
};

const getrandomOwner = (): string => {
    return owners[getrandomInt(owners.length - 1)];
};

const getrandomAssetID = (assetIDs: string[], max: number, prefix = "asset"): string => {
    const assetId = `${prefix}${getrandomInt(max)}`;
    if (assetIDs.find((id) => id === assetId)) return getrandomAssetID(assetIDs, max);
    return assetId;
};

const getRandomString = (length: number): string => {
    let result = "";
    const characters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
    for (let index = 0; index < length; index++) {
        result += characters.charAt(Math.floor(Math.random() * characters.length));
    }
    return result;
};

const getrandomEmail = (): string => {
    const domain = domains[getrandomInt(domains.length - 1)];
    return `${getRandomString(10)}@${domain}`;
};

const getrandomPostalId = (postalIds: string[], max: number, prefix: string): string => {
    const assetId = `${prefix}${getrandomInt(max)}`;
    if (postalIds.find((id) => id === assetId)) return getrandomAssetID(postalIds, max, prefix);
    return assetId;
};

export {getrandomInt, getrandomPostalId};
