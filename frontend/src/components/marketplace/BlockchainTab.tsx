import { useState, useEffect } from 'react';
import { Button } from '../ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Badge } from '../ui/badge';
import { Input } from '../ui/input';
import { Loader, TrendingUp, Hash, Clock, CheckCircle, XCircle, Search, ExternalLink } from 'lucide-react';
import { toast } from 'sonner';
import { apiService } from '../../services/api';
import { Block, Transaction, TransactionStatus, TransactionType } from '../../types/api';
import { Label } from '../ui/label';

export function BlockchainTab() {
  const [blocks, setBlocks] = useState<Block[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchHash, setSearchHash] = useState('');
  const [selectedTransaction, setSelectedTransaction] = useState<Transaction | null>(null);

  useEffect(() => {
    loadBlockchainData();
  }, []);

  const loadBlockchainData = async () => {
    try {
      setLoading(true);
      const response = await apiService.getBlockchainLedger({ limit: 20 });
      setBlocks(response.blocks);
    } catch (error) {
      console.error('Failed to load blockchain data:', error);
      toast.error('Failed to load blockchain data');
    } finally {
      setLoading(false);
    }
  };

  const handleSearchTransaction = async () => {
    if (!searchHash.trim()) {
      toast.error('Please enter a transaction hash');
      return;
    }

    try {
      const response = await apiService.getTransactionStatus(searchHash);
      setSelectedTransaction(response.transaction);
    } catch (error) {
      console.error('Failed to find transaction:', error);
      toast.error('Transaction not found');
      setSelectedTransaction(null);
    }
  };

  const getTransactionStatusBadge = (status: TransactionStatus) => {
    const statusConfig = {
      pending: { variant: 'secondary' as const, text: 'Pending', icon: Clock },
      confirmed: { variant: 'default' as const, text: 'Confirmed', icon: CheckCircle },
      failed: { variant: 'destructive' as const, text: 'Failed', icon: XCircle },
      expired: { variant: 'destructive' as const, text: 'Expired', icon: XCircle },
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

  if (loading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Loader className="w-6 h-6 text-green-600 animate-spin" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h2 className="text-2xl font-bold">Blockchain Explorer</h2>
        <p className="text-muted-foreground">Explore the immutable transaction ledger</p>
      </div>

      {/* Search Transaction */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Search className="h-5 w-5" />
            Search Transaction
          </CardTitle>
          <CardDescription>
            Enter a transaction hash to view its details
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex gap-2">
            <Input
              placeholder="Enter transaction hash..."
              value={searchHash}
              onChange={(e) => setSearchHash(e.target.value)}
              className="flex-1"
            />
            <Button onClick={handleSearchTransaction} disabled={!searchHash.trim()}>
              Search
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Selected Transaction Details */}
      {selectedTransaction && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Hash className="h-5 w-5" />
              Transaction Details
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <Label className="text-sm font-medium">Hash</Label>
                <p className="text-sm font-mono bg-muted p-2 rounded">{selectedTransaction.hash}</p>
              </div>
              <div>
                <Label className="text-sm font-medium">Status</Label>
                <div className="mt-1">{getTransactionStatusBadge(selectedTransaction.status)}</div>
              </div>
              <div>
                <Label className="text-sm font-medium">Type</Label>
                <div className="mt-1">{getTransactionTypeBadge(selectedTransaction.tx_type)}</div>
              </div>
              <div>
                <Label className="text-sm font-medium">Amount</Label>
                <p className="text-sm">{selectedTransaction.amount} coins</p>
              </div>
              <div>
                <Label className="text-sm font-medium">Nonce</Label>
                <p className="text-sm">{selectedTransaction.nonce}</p>
              </div>
              <div>
                <Label className="text-sm font-medium">Block Height</Label>
                <p className="text-sm">{selectedTransaction.block_height || 'Pending'}</p>
              </div>
            </div>
            <div className="pt-4 border-t">
              <Label className="text-sm font-medium">Signature</Label>
              <p className="text-sm font-mono bg-muted p-2 rounded text-xs break-all">
                {selectedTransaction.signature}
              </p>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Blockchain Blocks */}
      <div>
        <h3 className="text-lg font-semibold mb-4">Recent Blocks</h3>
        <div className="space-y-4">
          {blocks?.map((block) => (
            <Card key={block.id} className="hover:shadow-md transition-shadow">
              <CardHeader>
                <div className="flex justify-between items-start">
                  <div>
                    <CardTitle className="text-lg">Block #{block.height}</CardTitle>
                    <CardDescription>
                      {new Date(block.timestamp).toLocaleString()}
                    </CardDescription>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge variant="outline">{block.transactions.length} transactions</Badge>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <div className="space-y-2">
                  <div className="flex justify-between items-center text-sm">
                    <span className="text-muted-foreground">Block Hash:</span>
                    <span className="font-mono">{formatHash(block.hash)}</span>
                  </div>
                  <div className="flex justify-between items-center text-sm">
                    <span className="text-muted-foreground">Previous Hash:</span>
                    <span className="font-mono">{formatHash(block.previous_hash)}</span>
                  </div>
                </div>

                {/* Transactions in Block */}
                {block.transactions.length > 0 && (
                  <div className="mt-4 pt-4 border-t">
                    <h4 className="text-sm font-medium mb-2">Transactions</h4>
                    <div className="space-y-2">
                      {block.transactions.slice(0, 3).map((tx) => (
                        <div key={tx.id} className="flex justify-between items-center text-xs">
                          <div className="flex items-center gap-2">
                            {getTransactionTypeBadge(tx.tx_type)}
                            <span className="font-mono">{formatHash(tx.hash)}</span>
                          </div>
                          <div className="flex items-center gap-2">
                            <span>{tx.amount} coins</span>
                            {getTransactionStatusBadge(tx.status)}
                          </div>
                        </div>
                      ))}
                      {block.transactions.length > 3 && (
                        <div className="text-xs text-muted-foreground">
                          +{block.transactions.length - 3} more transactions
                        </div>
                      )}
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          ))}
        </div>

        {blocks?.length === 0 && (
          <div className="text-center py-8">
            <TrendingUp className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <h3 className="text-lg font-semibold mb-2">No blocks found</h3>
            <p className="text-muted-foreground">
              The blockchain is empty or still initializing.
            </p>
          </div>
        )}
      </div>
    </div>
  );
} 