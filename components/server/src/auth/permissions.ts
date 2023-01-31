/**
 * Copyright (c) 2023 Gitpod GmbH. All rights reserved.
 * Licensed under the GNU Affero General Public License (AGPL).
 * See License.AGPL.txt in the project root for license information.
 */

import { v1 } from "@authzed/authzed-node";
import { ErrorCodes } from "@gitpod/gitpod-protocol/lib/messaging/error";
import { ResponseError } from "vscode-ws-jsonrpc";

const client = v1.NewClient("mytokenhere");

type ObjectType = "team" | "user";

// type ResourceType = "organization";

function obj(type: ObjectType, id: string): v1.ObjectReference {
    return v1.ObjectReference.create({
        objectType: type,
        objectId: id,
    });
}

function userSubject(id: string): v1.SubjectReference {
    return v1.SubjectReference.create({
        object: obj("user", id),
    });
}

// type TeamRole = "member" | "owner";

// type TeamPermission = "read_team";

interface PermissionChecker {
    check(userID: string, resourceID: string): Promise<void>;
}

class DeclarativePermission implements PermissionChecker {
    private resourceType: ObjectType;
    private action: string;

    constructor(resourceType: ObjectType, action: string) {
        this.resourceType = resourceType;
        this.action = action;
    }

    async check(userID: string, resourceID: string): Promise<void> {
        const req = v1.CheckPermissionRequest.create({
            subject: userSubject(userID),
            permission: this.action,
            resource: obj(this.resourceType, resourceID),
        });

        const response = await client.promises.checkPermission(req);
        if (response.permissionship === v1.CheckPermissionResponse_Permissionship.HAS_PERMISSION) {
            return;
        }

        throw new ResponseError(
            ErrorCodes.PERMISSION_DENIED,
            `User (ID: ${userID}) is not permitted to perform ${this.action} on resource ${this.resourceType} (ID: ${resourceID}).`,
        );
    }
}

class StaticPermission implements PermissionChecker {
    private action: string;
    private result: boolean;
    private resourceType: ObjectType;

    constructor(resourceType: ObjectType, action: string, result: boolean) {
        this.result = result;
        this.action = action;
        this.resourceType = resourceType;
    }

    async check(userID: string, resourceID: string): Promise<void> {
        if (this.result) {
            return;
        }

        throw new ResponseError(
            ErrorCodes.PERMISSION_DENIED,
            `User (ID: ${userID}) is not permitted to perform ${this.action} on resource ${this.resourceType} (ID: ${resourceID}).`,
        );
    }
}

// Anyone is able to create a new team.
export const CreateTeam = new StaticPermission("team", "create", true);

export const ReadTeam = new DeclarativePermission("team", "read");

// async function writeTeamRole(teamID: string, role: TeamRole, userID: string) {
//     const req = v1.WriteRelationshipsRequest.create({
//         updates: [
//             v1.RelationshipUpdate.create({
//                 relationship: v1.Relationship.create({
//                     resource: obj("team", teamID),
//                     relation: role,
//                     subject: v1.SubjectReference.create({
//                         object: obj("user", userID),
//                     }),
//                 }),
//                 operation: v1.RelationshipUpdate_Operation.CREATE,
//             }),
//         ],
//     });

//     return client.promises.writeRelationships(req);
// }

// async function checkTeamRead(teamID: string, userID: string) {
//     const req = v1.CheckPermissionRequest.create({
//         subject: userSubject(userID),
//         permission: "read_team",
//         resource: obj("team", teamID),
//     });

//     return client.promises.checkPermission(req);
// }
