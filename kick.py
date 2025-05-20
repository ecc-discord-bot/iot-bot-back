import discord
from discord.ext import commands

intent = discord.Intents.all()

client = commands.Bot("", intents=intent)
token = ""
message = "体験入部期間が終了したため、サーバーからキックしました。"


@client.command()
@commands.has_permissions(administrator=True)
async def kick(_, member: discord.Member):
    dm_channel = member.dm_channel
    if dm_channel is None:
        dm_channel = await member.create_dm()
    await dm_channel.send(message)
    await member.kick(reason=message)


if __name__ == "__main__":
    client.run(token)
