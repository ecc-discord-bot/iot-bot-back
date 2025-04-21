import discord

token = ""

intents = discord.Intents.all()

client = discord.Client(intents=intents)

guild_id = 1095933862657405038  # Replace with your guild ID
roles = [
    1096283209316175894,
    1096017500849848391,
    1096017336563142737,
    1096012253842657350,
]


@client.event
async def on_ready():
    guild = client.get_guild(guild_id)
    if guild is None:
        print("Guild not found.")
        return
    members = guild.members
    for member in members:
        for m_role in member.roles:
            for i, role in enumerate(roles[1:], start=1):
                if m_role.id == role:
                    n_role_id = roles[i - 1]
                    n_role = guild.get_role(n_role_id)
                    if n_role is None:
                        n_role = await guild.fetch_role(n_role_id)
                    n_roles = member.roles[:]
                    n_roles.remove(m_role)
                    n_roles.append(n_role)
                    print(n_roles)
                    await member.edit(roles=n_roles)
                    print(f"{member.display_name} {m_role.name} -> {n_role.name}")
                    break
            else:
                continue
            break

    print("done")


if __name__ == "__main__":
    client.run(token)
