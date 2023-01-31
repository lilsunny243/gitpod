/**
 * Copyright (c) 2023 Gitpod GmbH. All rights reserved.
 * Licensed under the GNU Affero General Public License (AGPL).
 * See License.AGPL.txt in the project root for license information.
 */

import { v1 } from "@authzed/authzed-node";

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

const READ_TEAM = permission("team", "read_team");
console.log(READ_TEAM);

type ResourceCheck = (userID: string, resourceID: string) => Promise<boolean>;

function permission(
    resource: ObjectType,
    perm: string,
): (userID: string, resourceID: string) => Promise<v1.CheckPermissionResponse> {
    return async (userID: string, resourceID: string) => {
        const req = v1.CheckPermissionRequest.create({
            subject: userSubject(userID),
            permission: perm,
            resource: obj(resource, resourceID),
        });

        return client.promises.checkPermission(req);
    };
}

export async function check(c: ResourceCheck, userID: string, resourceID: string) {
    const resp = await c(userID, resourceID);
    console.log(resp);
    return resp.permissionship === v1.CheckPermissionResponse_Permissionship.HAS_PERMISSION;
}

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
