using System;
using System.Collections.Concurrent;
using System.Threading.Tasks;
using System.Linq;
using cartservice.interfaces;
using Hipstershop;
using NLog;

namespace cartservice.cartstore
{
    internal class LocalCartStore : ICartStore
    {
        // Maps between user and their cart
        private ConcurrentDictionary<string, Hipstershop.Cart> userCartItems = new ConcurrentDictionary<string, Hipstershop.Cart>();
        private readonly Hipstershop.Cart emptyCart = new Hipstershop.Cart();

        private static Logger logger = LogManager.GetCurrentClassLogger();

        public Task InitializeAsync()
        {
            logger.Debug("Local Cart Store was initialized");

            return Task.CompletedTask;
        }

        public Task AddItemAsync(string userId, string productId, int quantity)
        {
            logger.Info($"AddItemAsync called with userId={userId}, productId={productId}, quantity={quantity}");
            var newCart = new Hipstershop.Cart
                {
                    UserId = userId,
                    Items = { new Hipstershop.CartItem { ProductId = productId, Quantity = quantity } }
                };
            userCartItems.AddOrUpdate(userId, newCart,
            (k, exVal) =>
            {
                // If the item exists, we update its quantity
                var existingItem = exVal.Items.SingleOrDefault(item => item.ProductId == productId);
                if (existingItem != null)
                {
                    existingItem.Quantity += quantity;
                }
                else
                {
                    exVal.Items.Add(new Hipstershop.CartItem { ProductId = productId, Quantity = quantity });
                }

                return exVal;
            });

            return Task.CompletedTask;
        }

        public Task EmptyCartAsync(string userId)
        {
            logger.Info($"EmptyCartAsync called with userId={userId}");
            userCartItems[userId] = new Hipstershop.Cart();

            return Task.CompletedTask;
        }

        public Task<Hipstershop.Cart> GetCartAsync(string userId)
        {
            logger.Info($"GetCartAsync called with userId={userId}");
            Hipstershop.Cart cart = null;
            if (!userCartItems.TryGetValue(userId, out cart))
            {
                logger.Error($"No carts for user {userId}");
                return Task.FromResult(emptyCart);
            }

            return Task.FromResult(cart);
        }

        public bool Ping()
        {
            return true;
        }
    }
}