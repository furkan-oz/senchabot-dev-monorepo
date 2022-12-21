const { Permissions } = require("discord.js");
const { checkBotPermission } = require("./botFunctions");

function addRole(member, memberRole) {
  // Check if the bot has the Manage Roles permission.
  if (checkBotPermission(member.guild, Permissions.FLAGS.MANAGE_ROLES)) {
    console.log(`ADD "${memberRole.name}" ROLE -> "${member.displayName}".`);
    member.roles.add(memberRole).catch(console.error);
  } else {
    console.log(`BOT DOES NOT HAVE PERMISSIONS TO "MANAGE ROLES".`);
  }
}

function removeRole(member, memberRole) {
  // Check if the bot has the Manage Roles permission.
  if (checkBotPermission(member.guild, Permissions.FLAGS.MANAGE_ROLES)) {
    console.log(
      `REMOVE "${memberRole.name}" ROLE FROM "${member.displayName}".`
    );
    member.roles.remove(memberRole).catch(console.error);
  } else {
    console.log(`BOT DOES NOT HAVE PERMISSIONS TO "MANAGE ROLES".`);
  }
}

function addRoleAll(guild, memberRole) {
  guild.members.cache.forEach((member) => {
    const hasRole = member.roles.cache.some(
      (role) => role.id === memberRole.id
    );

    if (!hasRole && !member.user.bot) {
      addRole(member, memberRole);
    }
  });
}

function checkMemberPermission(memberPermissions, permissionFlag) {
  return memberPermissions.has(permissionFlag);
}

module.exports = { addRole, removeRole, addRoleAll, checkMemberPermission };
