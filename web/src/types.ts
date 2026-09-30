export type User={id:string;name:string;role:string};
export type Invitation={maxUses?:number;useCount?:number;workspaceTeamIds?:string[];id:string;code:string;role:string;targetUserId?:string;createdBy:string;created:string;expires:string;usedBy?:string;usedAt?:string;revokedAt?:string};
export type Team={reviewUntil?:string;parentTeamId?:string;id:string;name:string;description:string;color:string;members:User[];invites?:Invitation[];inviteCode?:string;inviteExpires?:string};
export type OrgTeam={id:string;parentTeamId?:string;name:string;memberCount:number};
export type Channel={id:string;teamId:string;name:string;description:string;color:string};
export type Comment={id:string;author:User;text:string;created:string;system:boolean};
export type ChannelMessage={id:string;channelId:string;author:User;text:string;created:string};
export type Version={id:string;name:string;mime:string;size:number;created:string;author:string};
export type Doc={importMethod?:string;externalId?:string;integrationId?:string;sourceSystem?:string;escalatedFrom?:string;escalatedRevision?:number;id:string;title:string;channelId:string;owner:User;status:string;due:string;description:string;created:string;updated:string;versions:Version[];comments:Comment[];revision:number};
export type Data={user:User;demo:boolean;demoProfiles?:User[];teams:Team[];orgChart:OrgTeam[];channels:Channel[];documents:Doc[];messages:ChannelMessage[];botConfigured:boolean};

export type FilePreview={kind:"document"|"spreadsheet";blocks?:{kind:string;text?:string;rows?:string[][]}[];sheets?:{name:string;rows:string[][]}[];truncated?:boolean};
