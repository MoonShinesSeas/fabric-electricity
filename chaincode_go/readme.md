1、钱包(Wallet)
Wallet（公钥哈希后的地址）
用户初始化钱包(Wallet)、将地址传入保存为钱包地址；

Monitor私有数据库
用户提供认证，返回公钥和环公钥

2、商品（Goods）
ID、Owner、数量、价格

3、订单（Order）
ID、价格密文、发送者钱包地址，接收者钱包地址(Sender\Reciver)、flag（0：未操作、1：同意、2：拒绝）

4、交易提案（Proposal）

Hello---(args:null){return "hello"}
InitLedger---(args:null){return "init success"}
SetWallet---(args:address(公钥哈希、钱包地址) string,ctext(公钥加密后的账户余额密文) string){return null}
GetWellt---(args:address(钱包地址) string){return:ctext}
GetAllGoods---(args:null){return:goods}
GetGoods---(args:id){return:good}
SetProposal---(args:proposalId,sender(发起者),reciver(接收者),ctext(价格密文))
GetProposal---(args:reciver(接收者),pri_str(私钥字符串)){return:proposals}
UpdateProposal---(args:reciver,sender,flag){return null}