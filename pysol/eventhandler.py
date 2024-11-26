import asyncio                                                                                                                                                                            
from telethon import TelegramClient, events                                                                                                                                               
                                                                                                                                                                                        
async def telegram():                                                                                                                                                                     
    api_id = 21465248  # api id                                                                                                                                                           
    api_hash = 'e6b7df34d0e42ca135ed3344bffec427'  # api hash                                                                                                                             
    phone_number = "5492235119311"  # phone number                                                                                                                                        
                                                                                                                                                                                        
    client = TelegramClient('anon', api_id, api_hash)                                                                                                                                     
                                                                                                                                                                                        
    @client.on(events.NewMessage(chats=-1002109566555))  # Escuchar el chat específico                                                                                                    
    async def my_event_handler(event):                                                                                                                                                    
        text = event.raw_text                                                                                                                                                             
        if "Platform: Raydium" in text:                                                                                                                                                   
            parts = text.split("Base: ")                                                                                                                                                  
            if len(parts) > 1:                                                                                                                                                            
                address = parts[1].split("\nQuote:")[0]                                                                                                                                   
                print("token:" + address)  # Enviar el address al stdout                                                                                                                             
                                                                                                                                                                                        
    await client.start(phone_number)                                                                                                                                                      
    await client.run_until_disconnected()                                                                                                                                                 
                                                                                                                                                                                        
if __name__ == "__main__":                                                                                                                                                                
    asyncio.run(telegram())
# // @fluxbot_pool_sniper || https://t.me/fluxbot_pool_sniper