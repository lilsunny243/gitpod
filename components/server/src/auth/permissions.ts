/**
 * Copyright (c) 2023 Gitpod GmbH. All rights reserved.
 * Licensed under the GNU Affero General Public License (AGPL).
 * See License.AGPL.txt in the project root for license information.
 */

import { v1 } from "@authzed/authzed-node";

const client = v1.NewClient("mytokenhere");

type ObjectType = "team" | "user";

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

type TeamRole = "member" | "owner";

type TeamPermission = "read_team";

async function writeTeamRole(teamID: string, role: TeamRole, userID: string) {
    const req = v1.WriteRelationshipsRequest.create({
        updates: [
            v1.RelationshipUpdate.create({
                relationship: v1.Relationship.create({
                    resource: obj("team", teamID),
                    relation: role,
                    subject: v1.SubjectReference.create({
                        object: obj("user", userID),
                    }),
                }),
                operation: v1.RelationshipUpdate_Operation.CREATE,
            }),
        ],
    });

    return client.promises.writeRelationships(req);
}

async function checkTeamRead(teamID: string, userID: string) {
    const req = v1.CheckPermissionRequest.create({
        subject: userSubject(userID),
        permission: "read_team",
        resource: obj("team", teamID),
    });

    return client.promises.checkPermission(req);
}
