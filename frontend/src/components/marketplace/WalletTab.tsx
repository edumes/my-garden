import { useState, useEffect } from 'react';
import { Button } from '../ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Badge } from '../ui/badge';
import { Input } from '../ui/input';
import { Label } from '../ui/label';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '../ui/dialog';
import { Loader, Coins, TrendingUp, Send, Shield, Key, Copy, CheckCircle } from 'lucide-react';
import { toast } from 'sonner';
import { apiService } from '../../services/api';
import { Wallet, Transaction, TransactionStatus, TransactionType } from '../../types/api';

interface WalletTabProps {
  wallet: Wallet | null;
  onRefresh: () => void;
}

export function WalletTab({ wallet, onRefresh }: WalletTabProps) {
  const [loading, setLoading] = useState(false);
  const [showTransferModal, setShowTransferModal] = useState(false);
  const [showCreateWalletModal, setShowCreateWalletModal] = useState(false);
  const [recentTransactions, setRecentTransactions] = useState<Transaction[]>([]);
  const [transferForm, setTransferForm] = useState({
    receiver_key: '',
    amount: ''
  });

  useEffect(() => {
    if (wallet) {
      loadRecentTransactions();
    }
  }, [wallet]);

  const loadRecentTransactions = async () => {
    try {
      const response = await apiService.getBlockchainLedger({ limit: 10 });
      // Filter transactions for current user (this would need backend support)
      setRecentTransactions(response.blocks.flatMap(block => block.transactions).slice(0, 10));
    } catch (error) {
      console.error('Failed to load recent transactions:', error);
    }
  };

  const handleCreateWallet = async () => {
    try {
      setLoading(true);
      // In a real implementation, you would generate a key pair here
      const encryptedPrivateKey = "mock_encrypted_private_key";
      await apiService.createWallet({ encrypted_private_key: encryptedPrivateKey });
      toast.success('Wallet created successfully');
      onRefresh();
      setShowCreateWalletModal(false);
    } catch (error) {
      console.error('Failed to create wallet:', error);
      toast.error('Failed to create wallet');
    } finally {
      setLoading(false);
    }
  };

  const handleTransfer = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!transferForm.receiver_key || !transferForm.amount) {
      toast.error('Please fill in all fields');
      return;
    }

    const amount = parseInt(transferForm.amount);
    if (isNaN(amount) || amount <= 0) {
      toast.error('Please enter a valid amount');
      return;
    }

    if (wallet && amount > wallet.balance) {
      toast.error('Insufficient balance');
      return;
    }

    try {
      setLoading(true);
      // In a real implementation, you would sign the transaction here
      const signature = "mock_signature";
      await apiService.transferCurrency({
        receiver_key: transferForm.receiver_key,
        amount,
        signature
      });
      
      toast.success('Transfer completed successfully');
      setTransferForm({ receiver_key: '', amount: '' });
      setShowTransferModal(false);
      onRefresh();
    } catch (error) {
      console.error('Failed to transfer currency:', error);
      toast.error('Failed to transfer currency');
    } finally {
      setLoading(false);
    }
  };

  const getTransactionStatusBadge = (status: TransactionStatus) => {
    const statusConfig = {
      pending: { variant: 'secondary' as const, text: 'Pending', icon: Loader },
      confirmed: { variant: 'default' as const, text: 'Confirmed', icon: CheckCircle },
      failed: { variant: 'destructive' as const, text: 'Failed', icon: CheckCircle },
      expired: { variant: 'destructive' as const, text: 'Expired', icon: CheckCircle },
    };
    const config = statusConfig[status];
    const Icon = config.icon;
    return (
      <Badge variant={config.variant} className="flex items-center gap-1">
        <Icon className="h-3 w-3" />
        {config.text}
      </Badge>
    );
  };

  const getTransactionTypeBadge = (type: TransactionType) => {
    const typeConfig = {
      delivery_reward: { variant: 'default' as const, text: 'Delivery Reward' },
      plant_sale: { variant: 'secondary' as const, text: 'Plant Sale' },
      transfer: { variant: 'outline' as const, text: 'Transfer' },
      commission: { variant: 'destructive' as const, text: 'Commission' },
    };
    const config = typeConfig[type];
    return (
      <Badge variant={config.variant} className="text-xs">
        {config.text}
      </Badge>
    );
  };

  const formatHash = (hash: string) => {
    return `${hash.substring(0, 8)}...${hash.substring(hash.length - 8)}`;
  };

  if (!wallet) {
    return (
      <div className="space-y-6">
        <div>
          <h2 className="text-2xl font-bold">Wallet</h2>
          <p className="text-muted-foreground">Create your blockchain wallet to start trading</p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Key className="h-5 w-5" />
              Create Wallet
            </CardTitle>
            <CardDescription>
              You need a wallet to participate in the marketplace
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button onClick={() => setShowCreateWalletModal(true)}>
              Create New Wallet
            </Button>
          </CardContent>
        </Card>

        {/* Create Wallet Modal */}
        <Dialog open={showCreateWalletModal} onOpenChange={setShowCreateWalletModal}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create New Wallet</DialogTitle>
              <DialogDescription>
                This will create a new blockchain wallet for your account. 
                Your private key will be encrypted and stored securely.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4">
              <div className="p-4 bg-muted rounded-lg">
                <h4 className="font-medium mb-2">Security Notice</h4>
                <ul className="text-sm text-muted-foreground space-y-1">
                  <li>• Your private key will be encrypted with AES-256-GCM</li>
                  <li>• Never share your private key with anyone</li>
                  <li>• Keep your account secure with a strong password</li>
                  <li>• All transactions are cryptographically signed</li>
                </ul>
              </div>
              <div className="flex justify-end gap-2">
                <Button
                  variant="outline"
                  onClick={() => setShowCreateWalletModal(false)}
                  disabled={loading}
                >
                  Cancel
                </Button>
                <Button onClick={handleCreateWallet} disabled={loading}>
                  {loading && <Loader className="w-4 h-4 mr-2 animate-spin" />}
                  Create Wallet
                </Button>
              </div>
            </div>
          </DialogContent>
        </Dialog>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h2 className="text-2xl font-bold">Wallet</h2>
        <p className="text-muted-foreground">Manage your blockchain wallet and transactions</p>
      </div>

      {/* Wallet Overview */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Balance</CardTitle>
            <Coins className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{wallet.balance.toLocaleString()}</div>
            <p className="text-xs text-muted-foreground">Available coins</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Reputation</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{wallet.reputation}</div>
            <p className="text-xs text-muted-foreground">Trust score</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Nonce</CardTitle>
            <Shield className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{wallet.nonce}</div>
            <p className="text-xs text-muted-foreground">Transaction counter</p>
          </CardContent>
        </Card>
      </div>

      {/* Wallet Actions */}
      <div className="flex gap-4">
        <Button onClick={() => setShowTransferModal(true)} className="flex items-center gap-2">
          <Send className="h-4 w-4" />
          Send Coins
        </Button>
        <Button variant="outline" className="flex items-center gap-2">
          <Copy className="h-4 w-4" />
          Copy Public Key
        </Button>
      </div>

      {/* Public Key */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Key className="h-5 w-5" />
            Public Key
          </CardTitle>
          <CardDescription>
            Share this public key to receive coins from other users
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-2">
            <code className="flex-1 p-2 bg-muted rounded text-sm break-all">
              {wallet.public_key}
            </code>
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                navigator.clipboard.writeText(wallet.public_key);
                toast.success('Public key copied to clipboard');
              }}
            >
              <Copy className="h-4 w-4" />
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Recent Transactions */}
      <div>
        <h3 className="text-lg font-semibold mb-4">Recent Transactions</h3>
        <div className="space-y-2">
          {recentTransactions.map((tx) => (
            <Card key={tx.id} className="hover:shadow-md transition-shadow">
              <CardContent className="p-4">
                <div className="flex justify-between items-center">
                  <div className="flex items-center gap-3">
                    {getTransactionTypeBadge(tx.tx_type)}
                    <div>
                      <p className="font-medium">{tx.amount} coins</p>
                      <p className="text-sm text-muted-foreground font-mono">
                        {formatHash(tx.hash)}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {getTransactionStatusBadge(tx.status)}
                    <span className="text-sm text-muted-foreground">
                      {new Date(tx.created_at).toLocaleDateString()}
                    </span>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
          
          {recentTransactions.length === 0 && (
            <div className="text-center py-8">
              <Coins className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <h3 className="text-lg font-semibold mb-2">No transactions yet</h3>
              <p className="text-muted-foreground">
                Start trading to see your transaction history here.
              </p>
            </div>
          )}
        </div>
      </div>

      {/* Transfer Modal */}
      <Dialog open={showTransferModal} onOpenChange={setShowTransferModal}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Send Coins</DialogTitle>
            <DialogDescription>
              Transfer coins to another user's wallet
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleTransfer} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="receiver_key">Receiver's Public Key</Label>
              <Input
                id="receiver_key"
                placeholder="Enter receiver's public key"
                value={transferForm.receiver_key}
                onChange={(e) => setTransferForm(prev => ({ ...prev, receiver_key: e.target.value }))}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="amount">Amount (coins)</Label>
              <Input
                id="amount"
                type="number"
                placeholder="Enter amount"
                value={transferForm.amount}
                onChange={(e) => setTransferForm(prev => ({ ...prev, amount: e.target.value }))}
                min="1"
                max={wallet.balance}
              />
              <p className="text-xs text-muted-foreground">
                Available: {wallet.balance} coins
              </p>
            </div>
            <div className="flex justify-end gap-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => setShowTransferModal(false)}
                disabled={loading}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={loading}>
                {loading && <Loader className="w-4 h-4 mr-2 animate-spin" />}
                Send Coins
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
} 