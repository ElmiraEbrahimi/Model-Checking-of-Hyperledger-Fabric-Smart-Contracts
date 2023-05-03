/*
  SPDX-License-Identifier: Apache-2.0
*/

import {Object, Property} from "fabric-contract-api";

@Object()
export class PostOffice {
    @Property()
    public name: string;

    @Property()
    public addresses: string[] = [];
}