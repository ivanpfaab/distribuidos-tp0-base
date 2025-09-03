import logging
from common.utils import load_bets, has_won


class Lottery:
    def __init__(self):
        self._lottery_conducted = False
        self._winners_cache = {}  # Cache winners by agency

    def conduct_lottery(self):
        """Conduct the lottery and find winners for each agency"""
        try:
            logging.info('action: sorteo | result: success')
            
            # Load all bets and find winners
            all_bets = list(load_bets())
            winners_by_agency = {}
            
            for bet in all_bets:
                if has_won(bet):
                    if bet.agency not in winners_by_agency:
                        winners_by_agency[bet.agency] = []
                    winners_by_agency[bet.agency].append(bet.document)
            
            # Cache winners by agency
            self._winners_cache = winners_by_agency
            self._lottery_conducted = True
            
            logging.info(f'action: lottery_winners_found | result: success | total_winners: {sum(len(winners) for winners in winners_by_agency.values())}')
            
        except Exception as e:
            logging.error(f'action: conduct_lottery | result: fail | error: {e}')

    def get_winners_for_agency(self, agency_id):
        """Get winners for a specific agency"""
        if agency_id in self._winners_cache:
            return self._winners_cache[agency_id]
        else:
            return []

    def is_lottery_conducted(self):
        """Check if lottery has been conducted"""
        return self._lottery_conducted